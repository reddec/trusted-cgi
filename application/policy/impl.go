package policy

import (
	"fmt"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/types"
	"sync"
)

// Store contains policies configuration for reload
type Store interface {
	// Save policies list
	SetPolicies(policies []application.Policy) error
	// Load policies list
	GetPolicies() ([]application.Policy, error)
}

func New(store Store) (*policiesImpl, error) {
	impl := &policiesImpl{
		store:            store,
		policiesByID:     map[string]*application.Policy{},
		policiesByLambda: map[string]string{},
	}
	return impl, impl.load()
}

type policiesImpl struct {
	store            Store
	lock             sync.RWMutex
	policiesByID     map[string]*application.Policy
	policiesByLambda map[string]string
}

func (policies *policiesImpl) load() error {
	list, err := policies.store.GetPolicies()
	if err != nil {
		return err
	}
	for _, item := range list {
		policies.policiesByID[item.ID] = &item
		for lambda := range item.Lambdas {
			policies.policiesByLambda[lambda] = item.ID
		}
	}
	return nil
}

func (policies *policiesImpl) List() []application.Policy {
	policies.lock.RLock()
	defer policies.lock.RUnlock()
	return policies.unsafeList()
}

func (policies *policiesImpl) Create(policy string, definition application.PolicyDefinition) (*application.Policy, error) {
	policies.lock.Lock()
	defer policies.lock.Unlock()
	_, exist := policies.policiesByID[policy]
	if exist {
		return nil, fmt.Errorf("policy %s already exists", policy)
	}
	info := &application.Policy{
		ID:         policy,
		Definition: definition,
		Lambdas:    make(types.JsonStringSet),
	}
	policies.policiesByID[policy] = info
	return info, policies.store.SetPolicies(policies.unsafeList())
}

func (policies *policiesImpl) Remove(policy string) error {
	policies.lock.Lock()
	defer policies.lock.Unlock()
	info, exist := policies.policiesByID[policy]
	if !exist {
		return fmt.Errorf("policy %s does not exists", policy)
	}
	for lambda := range info.Lambdas {
		delete(policies.policiesByLambda, lambda)
	}
	delete(policies.policiesByID, policy)
	return policies.store.SetPolicies(policies.unsafeList())
}

func (policies *policiesImpl) Update(policy string, definition application.PolicyDefinition) error {
	policies.lock.Lock()
	defer policies.lock.Unlock()
	info, exist := policies.policiesByID[policy]
	if !exist {
		return fmt.Errorf("policy %s does not exists", policy)
	}
	info.Definition = definition
	return policies.store.SetPolicies(policies.unsafeList())
}

func (policies *policiesImpl) Apply(lambda string, policy string) error {
	policies.lock.Lock()
	defer policies.lock.Unlock()
	info, exists := policies.policiesByID[policy]
	if !exists {
		return fmt.Errorf("policy %s does not exist", policy)
	}
	if info.Lambdas.Has(lambda) {
		// already applied
		return nil
	}
	policies.unsafeUnlink(lambda)
	info.Lambdas.Set(lambda)
	policies.policiesByLambda[lambda] = policy
	return policies.store.SetPolicies(policies.unsafeList())
}

func (policies *policiesImpl) Inspect(lambda string, request *types.Request) error {
	policy, applicable, err := policies.findPolicy(lambda)
	if err != nil {
		return err
	}
	if !applicable {
		return nil
	}
	return checkPolicy(policy, request)
}

func (policies *policiesImpl) Clear(lambda string) error {
	policies.lock.Lock()
	defer policies.lock.Unlock()
	if !policies.unsafeUnlink(lambda) {
		return nil
	}
	return policies.store.SetPolicies(policies.unsafeList())
}

func (policies *policiesImpl) unsafeUnlink(lambda string) bool {
	policyId, hasPolicy := policies.policiesByLambda[lambda]
	if !hasPolicy {
		return false
	}
	// remove direct ref
	delete(policies.policiesByLambda, lambda)

	// remove back ref
	if policy, exist := policies.policiesByID[policyId]; exist {
		policy.Lambdas.Del(lambda)
	}
	return true
}

func (policies *policiesImpl) unsafeList() []application.Policy {
	var ans = make([]application.Policy, 0, len(policies.policiesByID))
	for _, policy := range policies.policiesByID {
		ans = append(ans, *policy)
	}
	return ans
}

func (policies *policiesImpl) findPolicy(lambda string) (policy application.PolicyDefinition, applicable bool, err error) {
	policies.lock.RLock()
	defer policies.lock.RUnlock()
	policyId, exists := policies.policiesByLambda[lambda]
	if !exists {
		applicable = false
		return // no applied policy
	}
	info, exists := policies.policiesByID[policyId]
	if !exists {
		err = fmt.Errorf("corrupted policy data: lambda %s linked to unknown policy %s", lambda, policyId)
		return
	}
	return info.Definition, true, nil
}
