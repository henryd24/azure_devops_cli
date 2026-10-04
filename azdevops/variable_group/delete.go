package variable_group

import (
	"azuredevops/azdevops"
	"fmt"
	"net/url"
)

// DeleteVariableGroup elimina un Variable Group del proyecto indicado (projectID es el GUID).
func DeleteVariableGroup(client *azdevops.Client, id int, projectID string) error {
	q := url.Values{"api-version": {apiVersion}, "projectIds": {projectID}}
	if err := client.Do("DELETE", client.OrgURL(fmt.Sprintf("distributedtask/variablegroups/%d", id), q), nil, nil); err != nil {
		return fmt.Errorf("error al eliminar el Variable Group %d: %w", id, err)
	}
	return nil
}
