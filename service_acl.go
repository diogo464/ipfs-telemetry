package telemetry

import "github.com/libp2p/go-libp2p/core/peer"

type serviceAccessControl struct {
	accessType ServiceAccessType
	whitelist  map[peer.ID]struct{}
}

func newServiceAccessControl(accessType ServiceAccessType, whitelist map[peer.ID]struct{}) *serviceAccessControl {
	return &serviceAccessControl{
		accessType: accessType,
		whitelist:  whitelist,
	}
}

func (s *serviceAccessControl) isAllowed(id peer.ID) bool {
	switch s.accessType {
	case ServiceAccessPublic:
		return true
	case ServiceAccessRestricted:
		_, ok := s.whitelist[id]
		return ok
	default:
		return false
	}
}
