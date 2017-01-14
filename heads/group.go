package heads

import (
	"github.com/qmsk/go-web"
)

// Config
type GroupID string

type GroupConfig struct {
	Heads []HeadID
	Name  string
}

// heads
type groupMap map[GroupID]*Group

type APIGroups map[GroupID]APIGroup

func (groupMap groupMap) makeAPI() APIGroups {
	apiGroups := make(APIGroups)

	for groupID, group := range groupMap {
		apiGroups[groupID] = group.makeAPI()
	}

	return apiGroups
}

func (groupMap groupMap) makeAPIList() (apiGroups []APIGroup) {
	for _, group := range groupMap {
		apiGroups = append(apiGroups, group.makeAPI())
	}

	return
}

func (groupMap groupMap) GetREST() (web.Resource, error) {
	return groupMap.makeAPI(), nil
}

func (groupMap groupMap) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return groupMap.makeAPIList(), nil
	default:
		return groupMap[GroupID(name)], nil
	}
}

// Group
type Group struct {
	id     GroupID
	config GroupConfig
	heads  headMap
	events *Events

	intensity *GroupIntensity
	color     *GroupColor
}

func (group *Group) addHead(head *Head) {
	group.heads[head.id] = head
}

// initialize group parameters from heads
func (group *Group) init() {
	if groupIntensity := group.makeIntensity(); groupIntensity.exists() {
		group.intensity = &groupIntensity
	}

	if groupColor := group.makeColor(); groupColor.exists() {
		group.color = &groupColor
	}
}

func (group *Group) makeIntensity() GroupIntensity {
	var groupIntensity = GroupIntensity{
		heads: make(map[HeadID]HeadIntensity),
	}

	for headID, head := range group.heads {
		if headIntensity := head.parameters.Intensity; headIntensity != nil {
			groupIntensity.heads[headID] = *headIntensity
		}
	}

	return groupIntensity
}

func (group *Group) makeColor() GroupColor {
	var groupColor = GroupColor{
		headColors: make(map[HeadID]HeadColor),
	}

	for headID, head := range group.heads {
		if headColor := head.parameters.Color; headColor != nil {
			groupColor.headColors[headID] = *headColor
		}
	}

	return groupColor
}

// Web API
type APIGroupParams struct {
	group     *Group
	Intensity *APIIntensity `json:",omitempty"`
	Color     *APIColor     `json:",omitempty"`
}

type APIGroup struct {
	GroupConfig
	ID    GroupID
	Heads []HeadID

	APIGroupParams
}

func (group *Group) makeAPIHeads() (heads []HeadID) {
	for headID, _ := range group.heads {
		heads = append(heads, headID)
	}
	return
}

func (group *Group) makeAPI() APIGroup {
	return APIGroup{
		GroupConfig: group.config,
		ID:          group.id,
		Heads:       group.makeAPIHeads(),
		APIGroupParams: APIGroupParams{
			group:     group,
			Intensity: group.intensity.makeAPI(),
			Color:     group.color.makeAPI(),
		},
	}
}

func (group *Group) GetREST() (web.Resource, error) {
	return group.makeAPI(), nil
}
func (group *Group) PostREST() (web.Resource, error) {
	return &APIGroupParams{group: group}, nil
}

func (apiGroupParams APIGroupParams) Apply() error {
	if apiGroupParams.Intensity != nil {
		if err := apiGroupParams.Intensity.initGroup(apiGroupParams.group.intensity); err != nil {
			return web.RequestError(err)
		} else if err := apiGroupParams.Intensity.Apply(); err != nil {
			return err
		}
	}

	if apiGroupParams.Color != nil {
		if err := apiGroupParams.Color.initGroup(apiGroupParams.group.color); err != nil {
			return web.RequestError(err)
		} else if err := apiGroupParams.Color.Apply(); err != nil {
			return err
		}
	}

	return nil
}

// Web API Events
func (group *Group) Apply() error {
	group.events.updateGroup(group.id, group.makeAPI())

	for headID, head := range group.heads {
		head.events.updateHead(headID, head.makeAPI())
	}

	return nil
}
