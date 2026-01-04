package mesh

import (
	"sync"
	"time"
)

// PersonalNetwork represents a sub-mesh group (personal network)
type PersonalNetwork struct {
	ID          string
	Name        string
	Owner       string
	CreatedAt   time.Time
	Members     map[string]*NetworkMember
	Policies    *NetworkPolicy
	mu          sync.RWMutex
}

// NetworkMember represents a member in a personal network
type NetworkMember struct {
	NodeID      string
	JoinedAt    time.Time
	HasInternet bool
	IsProxy     bool
}

// NetworkPolicy defines policies for a personal network
type NetworkPolicy struct {
	AllowInternet bool
	AllowProxy    bool
	MaxBandwidth  int64 // bytes per second
	TTL           int   // Time to live for packets
}

// NewPersonalNetwork creates a new personal network
func NewPersonalNetwork(id, name, owner string) *PersonalNetwork {
	return &PersonalNetwork{
		ID:        id,
		Name:      name,
		Owner:     owner,
		CreatedAt: time.Now(),
		Members:   make(map[string]*NetworkMember),
		Policies: &NetworkPolicy{
			AllowInternet: true,
			AllowProxy:    true,
			MaxBandwidth:  0, // unlimited
			TTL:           64,
		},
	}
}

// AddMember adds a member to the personal network
func (pn *PersonalNetwork) AddMember(member *NetworkMember) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	pn.Members[member.NodeID] = member
}

// RemoveMember removes a member from the personal network
func (pn *PersonalNetwork) RemoveMember(nodeID string) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	delete(pn.Members, nodeID)
}

// GetMember retrieves a member by node ID
func (pn *PersonalNetwork) GetMember(nodeID string) (*NetworkMember, bool) {
	pn.mu.RLock()
	defer pn.mu.RUnlock()
	member, exists := pn.Members[nodeID]
	return member, exists
}

// GetAllMembers returns all members in the personal network
func (pn *PersonalNetwork) GetAllMembers() []*NetworkMember {
	pn.mu.RLock()
	defer pn.mu.RUnlock()
	members := make([]*NetworkMember, 0, len(pn.Members))
	for _, member := range pn.Members {
		members = append(members, member)
	}
	return members
}

// IsMember checks if a node is a member of the personal network
func (pn *PersonalNetwork) IsMember(nodeID string) bool {
	pn.mu.RLock()
	defer pn.mu.RUnlock()
	_, exists := pn.Members[nodeID]
	return exists
}

// GetProxyPeers returns all members in the network that can act as proxies
func (pn *PersonalNetwork) GetProxyPeers() []*NetworkMember {
	pn.mu.RLock()
	defer pn.mu.RUnlock()
	proxies := make([]*NetworkMember, 0)
	for _, member := range pn.Members {
		if member.HasInternet && member.IsProxy {
			proxies = append(proxies, member)
		}
	}
	return proxies
}

// PersonalNetworkManager manages all personal networks
type PersonalNetworkManager struct {
	Networks map[string]*PersonalNetwork
	mu       sync.RWMutex
}

// NewPersonalNetworkManager creates a new personal network manager
func NewPersonalNetworkManager() *PersonalNetworkManager {
	return &PersonalNetworkManager{
		Networks: make(map[string]*PersonalNetwork),
	}
}

// GetNetworkCount returns the number of personal networks
func (pnm *PersonalNetworkManager) GetNetworkCount() int {
	pnm.mu.RLock()
	defer pnm.mu.RUnlock()
	return len(pnm.Networks)
}

// CreateNetwork creates a new personal network
func (pnm *PersonalNetworkManager) CreateNetwork(id, name, owner string) *PersonalNetwork {
	pnm.mu.Lock()
	defer pnm.mu.Unlock()
	network := NewPersonalNetwork(id, name, owner)
	pnm.Networks[id] = network
	return network
}

// GetNetwork retrieves a personal network by ID
func (pnm *PersonalNetworkManager) GetNetwork(id string) (*PersonalNetwork, bool) {
	pnm.mu.RLock()
	defer pnm.mu.RUnlock()
	network, exists := pnm.Networks[id]
	return network, exists
}

// DeleteNetwork deletes a personal network
func (pnm *PersonalNetworkManager) DeleteNetwork(id string) {
	pnm.mu.Lock()
	defer pnm.mu.Unlock()
	delete(pnm.Networks, id)
}

// GetNetworksByOwner retrieves all personal networks owned by a user
func (pnm *PersonalNetworkManager) GetNetworksByOwner(owner string) []*PersonalNetwork {
	pnm.mu.RLock()
	defer pnm.mu.RUnlock()
	networks := make([]*PersonalNetwork, 0)
	for _, network := range pnm.Networks {
		if network.Owner == owner {
			networks = append(networks, network)
		}
	}
	return networks
}
