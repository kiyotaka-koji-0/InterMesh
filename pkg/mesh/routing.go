package mesh

import (
	"sync"
	"time"
)

// Route represents a route in the mesh network
type Route struct {
	Destination string
	NextHop     string
	HopCount    int
	Cost        int64 // metric for route selection
	LastUpdate  int64
}

// RoutingTable manages routes in the mesh network
type RoutingTable struct {
	Routes map[string]*Route
	mu     sync.RWMutex
}

// NewRoutingTable creates a new routing table
func NewRoutingTable() *RoutingTable {
	return &RoutingTable{
		Routes: make(map[string]*Route),
	}
}

// AddRoute adds a route to the routing table
func (rt *RoutingTable) AddRoute(destination, nextHop string, hopCount int, cost int64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.Routes[destination] = &Route{
		Destination: destination,
		NextHop:     nextHop,
		HopCount:    hopCount,
		Cost:        cost,
		LastUpdate:  getCurrentTimestamp(),
	}
}

// GetRoute retrieves a route from the routing table
func (rt *RoutingTable) GetRoute(destination string) (*Route, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	route, exists := rt.Routes[destination]
	return route, exists
}

// RemoveRoute removes a route from the routing table
func (rt *RoutingTable) RemoveRoute(destination string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	delete(rt.Routes, destination)
}

// UpdateRoute updates an existing route if the new one is better
func (rt *RoutingTable) UpdateRoute(destination, nextHop string, hopCount int, cost int64) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	
	existing, exists := rt.Routes[destination]
	if !exists || cost < existing.Cost {
		rt.Routes[destination] = &Route{
			Destination: destination,
			NextHop:     nextHop,
			HopCount:    hopCount,
			Cost:        cost,
			LastUpdate:  getCurrentTimestamp(),
		}
		return true
	}
	return false
}

// GetAllRoutes returns all routes in the routing table
func (rt *RoutingTable) GetAllRoutes() []*Route {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	routes := make([]*Route, 0, len(rt.Routes))
	for _, route := range rt.Routes {
		routes = append(routes, route)
	}
	return routes
}

// Router handles packet routing in the mesh network
type Router struct {
	NodeID       string
	RoutingTable *RoutingTable
	mu           sync.RWMutex
}

// NewRouter creates a new router
func NewRouter(nodeID string) *Router {
	return &Router{
		NodeID:       nodeID,
		RoutingTable: NewRoutingTable(),
	}
}

// RoutePacket routes a packet to its destination
func (r *Router) RoutePacket(destinationID string) (nextHop string, found bool) {
	route, exists := r.RoutingTable.GetRoute(destinationID)
	if !exists {
		return "", false
	}
	return route.NextHop, true
}

// UpdateRoutes updates routes based on peer information
func (r *Router) UpdateRoutes(peers []*Peer) {
	// Direct peers have hop count 1
	for _, peer := range peers {
		r.RoutingTable.UpdateRoute(peer.NodeID, peer.NodeID, 1, int64(calculateCost(peer)))
	}
}

// calculateCost calculates a route cost based on signal strength
func calculateCost(peer *Peer) int {
	// Higher RSSI is better, so we invert it
	// RSSI is typically negative, ranging from -30 (excellent) to -100 (poor)
	rssi := peer.RSSI
	if rssi > 0 {
		rssi = -rssi
	}
	// Convert to cost (lower is better): -30 becomes 30, -100 becomes 100
	return -rssi
}

// getCurrentTimestamp returns the current timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// UpdateRoute updates or adds a route in the routing table
func (r *Router) UpdateRoute(destination, nextHop string, hopCount int, latency time.Duration) {
	r.RoutingTable.AddRoute(destination, nextHop, hopCount, int64(latency.Milliseconds()))
}

// RemoveRoute removes a route from the routing table
func (r *Router) RemoveRoute(destination string) {
	r.RoutingTable.RemoveRoute(destination)
}

// GetRoute retrieves a route for a destination
func (r *Router) GetRoute(destination string) *Route {
	if route, exists := r.RoutingTable.GetRoute(destination); exists {
		return route
	}
	return nil
}
