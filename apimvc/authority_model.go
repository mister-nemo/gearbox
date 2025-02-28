package apimvc

import (
	"fmt"
	"gearbox/apiworks"
	"gearbox/types"
	"github.com/gearboxworks/go-status"
	"github.com/gearboxworks/go-status/only"
	"strings"
)

const AuthorityModelType = "authority"

var NilAuthorityModel = (*AuthorityModel)(nil)
var _ ItemModeler = NilAuthorityModel

type AuthorityModelMap map[types.Stackname]*AuthorityModel
type AuthorityModels []*AuthorityModel

type AuthorityModel struct {
	AuthorityId types.AuthorityDomain `json:"authority_id"`
	Model
}

func (me *AuthorityModel) GetAttributeMap() apiworks.AttributeMap {
	panic("implement me")
}

func NewFromGearsAuthority(ctx *Context, authority types.AuthorityDomain) (gs *AuthorityModel, sts Status) {
	return NewAuthority(authority), sts
}

func NewAuthority(authority types.AuthorityDomain) *AuthorityModel {
	return &AuthorityModel{
		AuthorityId: authority,
	}
}

func (me *AuthorityModel) GetType() ItemType {
	return AuthorityModelType
}

func (me *AuthorityModel) GetFullStackname() types.Stackname {
	return types.Stackname(me.GetId())
}

func (me *AuthorityModel) GetId() ItemId {
	return ItemId(me.AuthorityId)
}

func (me *AuthorityModel) SetId(itemid ItemId) (sts Status) {
	for range only.Once {
		if !strings.Contains(string(itemid), ".") {
			sts = status.Fail(&status.Args{
				Message: fmt.Sprintf("authority domain '%s' does not contain a period ('.')", itemid),
			})
			break
		}
		me.AuthorityId = types.AuthorityDomain(itemid)
	}
	return sts
}
