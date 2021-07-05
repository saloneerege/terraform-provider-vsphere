# vsphere_namespace

The `vsphere_namespace` resource cane be used to create namespaces on a Supervisor Cluster and configure them with resource quotas, storage, as well as set permissions for DevOps engineer users.

For more information on Namespaces, see [this
page][ref-vsphere-namespace].

[ref-vsphere-namespace]: https://docs.vmware.com/en/VMware-vSphere/7.0/vmware-vsphere-with-tanzu/GUID-1544C9FE-0B23-434E-B823-C59EFC2F7309.html

## Example Usage

The basic example below sets up a vApp container and a virtual machine in a
compute cluster and then creates a vApp entity to change the virtual machine's
power on behavior in the vApp container.

```hcl
data "vsphere_datacenter" "dc" {
  name = "dc1"
}

data "vsphere_compute_cluster" "compute_cluster"{
   name = "cluster-1"
   datacenter_id = data.vsphere_datacenter.dc.id
}

resource "vsphere_namespace" "namespace1"{
  cluster=data.vsphere_compute_cluster.compute_cluster.id
  namespace="terraform-test-namespace"
  description ="Testing Namespace resource creation"
  access_list {
      role = "EDIT"
      subject_type = "USER"
      subject = "user1"
      domain = "domain1"
  }
  storage_specifications {
      policy = data.vsphere_storage_policy.policy1.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `namespace` - (Required) Identifier of the namespace. This has DNS_LABEL restrictions as specified in . This must be an alphanumeric (a-z and 0-9) string and with maximum length of 63 characters and with the ‘-’ character allowed anywhere except the first or last character. This name is unique across all Namespaces in this vCenter server.
* `cluster` - (Required) The [managed object reference ID][docs-about-morefs] of the cluster on which the namespace is being created. When clients pass a value of this structure as a parameter, the field must be an identifier for the resource vsphere_compute_cluster. 
* `description` - (Optional) Description for the namespace. If unset, no description is added to the namespace.
* `access_list` - (Optional) Access controls associated with the namespace. If unset, only users with Administrator role can access the namespace.
* `storage_specifications` - (Optional) Storage associated with the namespace. If unset, storage policies will not be associated with the namespace which will prevent users from being able to provision pods with persistent storage on the namespace. Users will be able to provision pods which use local storage.


[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

## Attribute Reference

The following attributes are exported:

* `id`: The namespace is the ID.
* `configuration_status`: Describes the status of configuration for the namespace.CONFIGURING : The configuration is being applied to the namespace. REMOVING : The configuration is being removed and namespace is being deleted. RUNNING : The namespace is configured correctly. ERROR : Failed to apply the configuration to the namespace, user intervention needed.
*`instance_stats` :The basic runtime statistics about the namespace.
