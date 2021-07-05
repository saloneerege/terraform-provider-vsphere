package vsphere

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"gitlab.eng.vmware.com/openapi-sdks/vmware-infrastructure-sdk-go/runtime"
	"gitlab.eng.vmware.com/openapi-sdks/vmware-infrastructure-sdk-go/services/vsphere/vcenter/namespaces"
)

var accessRoleAllowedValues = []string{
	string(namespaces.NAMESPACESACCESSROLE_EDIT),
	string(namespaces.NAMESPACESACCESSROLE_VIEW),
}

var accessSubjectTypeAllowedValues = []string{
	string(namespaces.NAMESPACESACCESSSUBJECTTYPE_USER),
	string(namespaces.NAMESPACESACCESSSUBJECTTYPE_GROUP),
}

func resourceVsphereNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceVsphereNamespaceCreate,
		Read:   resourceVsphereNamespaceRead,
		Update: resourceVsphereNamespaceUpdate,
		Delete: resourceVsphereNamespaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Identifier of the namespace. This has DNS_LABEL restrictions as specified in . This must be an alphanumeric (a-z and 0-9) string and with maximum length of 63 characters and with the ‘-’ character allowed anywhere except the first or last character. This name is unique across all Namespaces in this vCenter server.",
			},
			"cluster": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Identifier of the cluster on which the namespace is being created. When clients pass a value of this structure as a parameter, the field must be an identifier for the resource vsphere_compute_cluster",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description for the namespace. If unset, no description is added to the namespace.",
			},
			"access_list": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subject": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the subject.",
						},
						"subject_type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Type of the subject.",
							ValidateFunc: validation.StringInSlice(accessSubjectTypeAllowedValues, false),
						},
						"domain": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Domain of the subject.",
						},
						"role": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Role of the subject on the namespace instance.",
							ValidateFunc: validation.StringInSlice(accessRoleAllowedValues, false),
						},
					},
				},
				Description: "Access controls associated with the namespace. If unset, only users with Administrator role can access the namespace.",
				Optional:    true,
			},
			"storage_specifications": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the storage policy. A Kubernetes storage class is created for this storage policy if it does not exist already. When clients pass a value of this structure as a parameter, the field must be an identifier for the resource type: SpsStorageProfile. When operations return a value of this structure as a result, the field will be an identifier for the resource type: SpsStorageProfile.",
						},
						"limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The maximum amount of storage (in mebibytes) which can be utilized by the namespace for this specification. If unset, no limits are placed.",
						},
					},
				},
				Description: "Storage associated with the namespace. If unset, storage policies will not be associated with the namespace which will prevent users from being able to provision pods with persistent storage on the namespace. Users will be able to provision pods which use local storage.",
				Optional:    true,
			},
			"configuration_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Describes the status of configuration for the namespace. CONFIGURING : The configuration is being applied to the namespace. REMOVING : The configuration is being removed and namespace is being deleted. RUNNING : The namespace is configured correctly. ERROR : Failed to apply the configuration to the namespace, user intervention needed.",
			},
			"instance_stats": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The basic runtime statistics about the namespace.",
			},
		},
	}
}

func resourceVsphereNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	APIClient := meta.(*VSphereClient).apiClient
	namespaceConfig := GetNamespaceRuntimeConfig(APIClient.SessionID, APIClient.BasePath, APIClient.InsecureFlag)

	nameSpace := d.Get("namespace").(string)
	clusterID := d.Get("cluster").(string)
	description := d.Get("description").(string)
	accessList := flattenAccessList(d.Get("access_list").([]interface{}))
	storageSpecsList := flattenStorageSpecifications(d.Get("storage_specifications").([]interface{}))

	ctx := context.WithValue(context.Background(), runtime.ContextAPIKey, runtime.APIKey{
		Key:    APIClient.SessionID,
		Prefix: "",
	})

	client := namespaces.NewAPIClient(namespaceConfig)

	nameSpaceCreate := namespaces.NamespacesInstancesCreateSpec{Cluster: clusterID, Namespace: nameSpace, Description: description, AccessList: accessList, StorageSpecs: storageSpecsList}

	response, err := client.InstancesApi.Create(ctx, nameSpaceCreate)
	if err != nil && response.StatusCode != 204 {
		return fmt.Errorf("error while creating namespace :%v", err)
	}

	d.SetId(nameSpace)
	return resourceVsphereNamespaceRead(d, meta)
}

func resourceVsphereNamespaceRead(d *schema.ResourceData, meta interface{}) error {

	APIClient := meta.(*VSphereClient).apiClient
	namespaceConfig := GetNamespaceRuntimeConfig(APIClient.SessionID, APIClient.BasePath, APIClient.InsecureFlag)

	nameSpace := d.Id()

	ctx := context.WithValue(context.Background(), runtime.ContextAPIKey, runtime.APIKey{
		Key:    APIClient.SessionID,
		Prefix: "",
	})

	client := namespaces.NewAPIClient(namespaceConfig)
	namespaceInfo, _, err := client.InstancesApi.Get(ctx, nameSpace)
	if err != nil {
		return fmt.Errorf("error while reading namespace :%v", err)
	}

	d.Set("configuration_status", namespaceInfo.ConfigStatus)
	d.Set("namespace", nameSpace)
	d.Set("cluster", namespaceInfo.Cluster)
	d.Set("description", namespaceInfo.Description)

	instanceStats := map[string]int64{}
	instanceStats["cpu_used"] = namespaceInfo.Stats.CpuUsed
	instanceStats["memory_used"] = namespaceInfo.Stats.MemoryUsed
	instanceStats["storage_used"] = namespaceInfo.Stats.StorageUsed
	log.Printf("Instance Stats :%v", instanceStats)
	d.Set("instance_stats", instanceStats)
	return nil
}

func resourceVsphereNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	APIClient := meta.(*VSphereClient).apiClient
	namespaceConfig := GetNamespaceRuntimeConfig(APIClient.SessionID, APIClient.BasePath, APIClient.InsecureFlag)

	client := namespaces.NewAPIClient(namespaceConfig)
	nameSpace := d.Get("namespace").(string)

	ctx := context.WithValue(context.Background(), runtime.ContextAPIKey, runtime.APIKey{
		Key:    APIClient.SessionID,
		Prefix: "",
	})

	description := d.Get("description").(string)
	accessList := flattenAccessList(d.Get("access_list").([]interface{}))
	storageSpecsList := flattenStorageSpecifications(d.Get("storage_specifications").([]interface{}))
	updateSpec := namespaces.NamespacesInstancesUpdateSpec{Description: description, AccessList: accessList, StorageSpecs: storageSpecsList}

	_, err := client.InstancesApi.Update(ctx, nameSpace, updateSpec)
	if err != nil {
		return fmt.Errorf("error while updating namespace :%v", err)
	}
	return nil
}

func resourceVsphereNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	APIClient := meta.(*VSphereClient).apiClient
	namespaceConfig := GetNamespaceRuntimeConfig(APIClient.SessionID, APIClient.BasePath, APIClient.InsecureFlag)
	nameSpace := d.Id()
	ctx := context.WithValue(context.Background(), runtime.ContextAPIKey, runtime.APIKey{
		Key:    APIClient.SessionID,
		Prefix: "",
	})

	client := namespaces.NewAPIClient(namespaceConfig)
	_, err := client.InstancesApi.Delete(ctx, nameSpace)
	if err != nil {
		return fmt.Errorf("error while deleting namespace :%v", err)
	}

	d.SetId("")
	return nil
}

func GetNamespaceRuntimeConfig(sessionID string, basePath string, insecureFlag bool) *runtime.Configuration {
	cfg := runtime.NewConfiguration()
	cfg.BasePath = basePath
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureFlag},
	}
	cfg.HTTPClient = &http.Client{Transport: tr}
	cfg.AddDefaultHeader("vmware-api-session-id", sessionID)
	return cfg
}

func flattenAccessList(accessListInterface []interface{}) []namespaces.NamespacesInstancesAccess {
	if len(accessListInterface) == 0 {
		return nil
	}
	var accessLists []namespaces.NamespacesInstancesAccess
	for _, accessList := range accessListInterface {
		a := accessList.(map[string]interface{})
		subjectType := a["subject_type"].(string)
		role := a["role"].(string)
		accessObj := namespaces.NamespacesInstancesAccess{Subject: a["subject"].(string), SubjectType: namespaces.NamespacesAccessSubjectType(subjectType), Domain: a["domain"].(string),
			Role: namespaces.NamespacesAccessRole(role)}

		accessLists = append(accessLists, accessObj)
	}
	return accessLists
}

func flattenStorageSpecifications(storageSpecificationsInterface []interface{}) []namespaces.NamespacesInstancesStorageSpec {
	if len(storageSpecificationsInterface) == 0 {
		return nil
	}
	var storageSpecsList []namespaces.NamespacesInstancesStorageSpec
	for _, storageSpecList := range storageSpecificationsInterface {
		s := storageSpecList.(map[string]interface{})
		limit := s["limit"].(int)
		storageSpecObj := namespaces.NamespacesInstancesStorageSpec{Policy: s["policy"].(string), Limit: int64(limit)}
		storageSpecsList = append(storageSpecsList, storageSpecObj)
	}
	return storageSpecsList
}
