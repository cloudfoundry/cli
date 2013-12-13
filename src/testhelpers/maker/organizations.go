/**
 * Created with IntelliJ IDEA.
 * User: pivotal
 * Date: 12/13/13
 * Time: 11:19 AM
 * To change this template use File | Settings | File Templates.
 */
package maker

import "cf"

var orgGuid func() string

func init() {
	orgGuid = guidGenerator("org")
}

func NewOrg(overrides Overrides) (org cf.Organization) {
	org.Name = "new-org"
	org.Guid = orgGuid()

	if overrides.Has("guid"){
		org.Guid = overrides["guid"].(string)
	}

	if overrides.Has("name") {
		org.Name = overrides["name"].(string)

	}

	return
}
