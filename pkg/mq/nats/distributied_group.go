package nats

import "strconv"

type _DistributedGroup string

func (group _DistributedGroup) String() string {
	return string(group)
}

func _GetDistributedGroupName(distribution int, distributedID int64) (int, _DistributedGroup) {
	mod := int(distributedID) % distribution
	return mod, _DistributedGroup(strconv.Itoa(mod))
}

func _GetGroups(distribution int) []_DistributedGroup {
	groups := make([]_DistributedGroup, distribution)
	for i := 0; i < distribution; i++ {
		groups[i] = _DistributedGroup(strconv.Itoa(i))
	}
	return groups
}
