package isolation

import (
	"context"
	"sync"
)

// AccountIsolation manages which accounts are available to which merchants.
type AccountIsolation struct {
	mu            sync.RWMutex
	groupAccounts map[string][]string // groupName -> accountIDs
	merchantGroup map[string]string   // merchantID -> groupName
}

func NewAccountIsolation() *AccountIsolation {
	return &AccountIsolation{
		groupAccounts: make(map[string][]string),
		merchantGroup: make(map[string]string),
	}
}

func (ai *AccountIsolation) AssignGroup(merchantID, groupName string) {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	ai.merchantGroup[merchantID] = groupName
}

func (ai *AccountIsolation) AddAccountToGroup(groupName, accountID string) {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	ai.groupAccounts[groupName] = append(ai.groupAccounts[groupName], accountID)
}

func (ai *AccountIsolation) RemoveAccountFromGroup(groupName, accountID string) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	accounts := ai.groupAccounts[groupName]
	for i, a := range accounts {
		if a == accountID {
			ai.groupAccounts[groupName] = append(accounts[:i], accounts[i+1:]...)
			return
		}
	}
}

func (ai *AccountIsolation) GetAccountsForMerchant(ctx context.Context, merchantID string) ([]string, error) {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	groupName, ok := ai.merchantGroup[merchantID]
	if !ok {
		return nil, nil
	}
	// Return a copy to prevent modification
	accounts := make([]string, len(ai.groupAccounts[groupName]))
	copy(accounts, ai.groupAccounts[groupName])
	return accounts, nil
}

func (ai *AccountIsolation) GetMerchantGroup(merchantID string) string {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	return ai.merchantGroup[merchantID]
}

func (ai *AccountIsolation) ListGroups() []string {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	groups := make([]string, 0, len(ai.groupAccounts))
	for g := range ai.groupAccounts {
		groups = append(groups, g)
	}
	return groups
}

func (ai *AccountIsolation) CountAccounts(groupName string) int {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	return len(ai.groupAccounts[groupName])
}
