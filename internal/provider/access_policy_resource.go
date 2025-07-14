package provider

import (
	"context"
	"encoding/binary"
	"fmt"
	"terraform-provider-izysafe/internal/client"

	"terraform-provider-izysafe/internal/proto/request"
	"terraform-provider-izysafe/internal/proto/response"

	"github.com/fxamacker/cbor/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/proto"
)

func resourceAccessPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccessPolicyCreate,
		ReadContext:   resourceAccessPolicyRead,
		DeleteContext: resourceAccessPolicyDelete,
		UpdateContext: resourceAccessPolicyUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: resourceMyBucketImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the folder in the root folder",
			},
			"max_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maimum size of the folder including all files and their versions",
			},
			"max_file_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum size of file that can be uploaded in the folder",
			},
			"max_file_versions": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of previous versions of each file to be stored in history as versions",
			},
			"remove_older_versions": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If true, remove the older versions as new versions are uploaded. Default true.",
			},
			"default_ttl_for_files": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Time(in s) for the file to be automatically deleted after the latest change.",
			},
		},
	}
}

func getMetaFrom(name string, client *client.Client) (response.GetMetaFromPath, diag.Diagnostics) {
	if client == nil {
		return response.GetMetaFromPath{}, diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	getMeta := request.GetMetaFromPath{
		Path:       "/" + name,
		Trashed:    false,
		TypeOfPath: 0,
	}
	req := request.Request{
		Operation: &request.Request_GetMetaFromPath{
			GetMetaFromPath: &getMeta,
		},
	}
	res, err := client.Send(&req)
	if err != nil {
		return response.GetMetaFromPath{}, diag.Errorf("Request/Response sent/recieved incorrectly" + err.Error())
	}
	if res == nil {
		return response.GetMetaFromPath{}, diag.Errorf("Empty Response")
	}
	return *res.GetGetMetaFromPath(), nil
}

func resourceMyBucketImportState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	name := d.Id()
	client := m.(*client.Client)
	if client == nil {
		return nil, fmt.Errorf("client is nil, please check the token and pin. contact support if the issue persists")
	}
	stat, err := getMetaFrom(name, client)
	if err != nil {
		return []*schema.ResourceData{d}, fmt.Errorf("%v", err)
	}

	if stat.Status == 0 {
		folderMeta := stat.Meta.GetFolderMeta()
		if folderMeta == nil {
			return []*schema.ResourceData{d}, fmt.Errorf("given name is not of a folder. read folder invalid")
		}
		d.Set("name", name)
		var policyObj request.Policy
		policyBytes := folderMeta.Policy
		err := proto.Unmarshal(policyBytes, &policyObj)
		if err != nil {
			return []*schema.ResourceData{d}, fmt.Errorf("wrong reponse sent. read folder failed")
		}
		attrList := policyObj.AttrToValue
		var keylist []string
		for _, attr := range attrList {
			keylist = append(keylist, attr.Attribute)
		}
		maxSize := 0
		maxFileSize := 0
		maxFileVersions := 0
		var autrotateVersions *bool
		var defaultTtlForFiles *uint64
		for idx, k := range keylist {
			if k == "max_size" {
				err := cbor.Unmarshal(attrList[idx].GetValue(), maxSize)
				if err != nil {
					return []*schema.ResourceData{d}, fmt.Errorf("data corrupted. read folder fialed")
				}
				break
			} else if k == "max_file_size" {
				err := cbor.Unmarshal(attrList[idx].GetValue(), maxFileSize)
				if err != nil {
					return []*schema.ResourceData{d}, fmt.Errorf("data corrupted. read folder fialed")
				}
				break
			} else if k == "max_file_versions" {
				err := cbor.Unmarshal(attrList[idx].GetValue(), maxFileVersions)
				if err != nil {
					return []*schema.ResourceData{d}, fmt.Errorf("data corrupted. read folder fialed")
				}
				break
			} else if k == "default_ttl_for_files" {
				defaultTtlForFiles = new(uint64)
				err := cbor.Unmarshal(attrList[idx].GetValue(), &defaultTtlForFiles)
				if err != nil {
					return []*schema.ResourceData{d}, fmt.Errorf("data corrupted. read folder fialed")
				}
				break
			} else if k == "remove_older_versions" {
				autrotateVersions = new(bool)
				err := cbor.Unmarshal(attrList[idx].GetValue(), &autrotateVersions)
				if err != nil {
					return []*schema.ResourceData{d}, fmt.Errorf("data corrupted. read folder fialed")
				}
				break
			}
		}
		if maxSize > 0 {
			d.Set("max_size", maxSize)
		}
		if maxFileSize > 0 {
			d.Set("max_file_size", maxFileSize)
		}
		if maxFileVersions > 0 {
			d.Set("max_file_versions", maxFileVersions)
		}
		if autrotateVersions != nil {
			d.Set("remove_older_versions", *autrotateVersions)
		}
		if defaultTtlForFiles != nil {
			d.Set("default_ttl_for_files", *defaultTtlForFiles)
		}
	} else {
		return []*schema.ResourceData{d}, fmt.Errorf("folder doesn't exists. import failed")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceAccessPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	client := m.(*client.Client)
	if client == nil {
		return diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	stat, err := getMetaFrom(name, client)
	if err != nil {
		return err
	}

	switch stat.Status {
	case response.Status_SUCCESS:
		return diag.Errorf("Folder already exists. Create not valid!!!")
	case response.Status_OBJECT_NOT_FOUND:
		createFold := request.CreateFolder{
			Name:       name,
			ParentPath: "/",
			TypeOfPath: 1,
		}
		req := request.Request{
			Operation: &request.Request_CreateFolder{
				CreateFolder: &createFold,
			},
		}
		resp, err := client.Send(&req)
		if err != nil {
			return diag.Errorf("Create Folder failed!!!")
		}
		if resp == nil {
			return diag.Errorf("Create Folder failed!!!")
		}
		if resp.GetCreateFolder().Status != 0 {
			return diag.Errorf("Create Folder failed!!!")
		}
	default:
		return diag.Errorf("Backend Error with status %s. Create Folder failed!!!", stat.Status)
	}
	d.SetId(name)

	return nil
}

func resourceAccessPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	client := m.(*client.Client)
	if client == nil {
		return diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	stat, err := getMetaFrom(name, client)
	if err != nil {
		return err
	}

	if stat.Status == 0 {
		folderMeta := stat.Meta.GetFolderMeta()
		if folderMeta == nil {
			return diag.Errorf("Given name is not of a folder. Read Folder Invalid!!!")
		}
		var policyObj request.Policy
		policyBytes := folderMeta.Policy
		err := proto.Unmarshal(policyBytes, &policyObj)
		if err != nil {
			return diag.Errorf("Wrong reponse sent. Read folder failed!!!")
		}
		attrList := policyObj.AttrToValue
		var keylist []string
		for _, attr := range attrList {
			keylist = append(keylist, attr.Attribute)
		}
		maxSize := 0
		maxFileSize := 0
		maxFileVersions := 0
		var autrotateVersions *bool
		var defaultTtlForFiles *uint64
		for idx, k := range keylist {
			if k == "max_size" {
				err := cbor.Unmarshal(attrList[idx].GetValue(), maxSize)
				if err != nil {
					return diag.Errorf("Data Corrupted. Read folder fialed!!!")
				}
				break
			} else if k == "max_file_size" {
				err := cbor.Unmarshal(attrList[idx].GetValue(), maxFileSize)
				if err != nil {
					return diag.Errorf("Data Corrupted. Read folder fialed!!!")
				}
				break
			} else if k == "max_file_versions" {
				err := cbor.Unmarshal(attrList[idx].GetValue(), maxFileVersions)
				if err != nil {
					return diag.Errorf("Data Corrupted. Read folder fialed!!!")
				}
				break
			} else if k == "default_ttl_for_files" {
				defaultTtlForFiles = new(uint64)
				err := cbor.Unmarshal(attrList[idx].GetValue(), &defaultTtlForFiles)
				if err != nil {
					return diag.Errorf("Data Corrupted. Read folder fialed!!!")
				}
				break
			} else if k == "remove_older_versions" {
				autrotateVersions = new(bool)
				err := cbor.Unmarshal(attrList[idx].GetValue(), &autrotateVersions)
				if err != nil {
					return diag.Errorf("Data Corrupted. Read folder fialed!!!")
				}
				break
			}
		}
		if maxSize > 0 {
			d.Set("max_size", maxSize)
		}
		if maxFileSize > 0 {
			d.Set("max_file_size", maxFileSize)
		}
		if maxFileVersions > 0 {
			d.Set("max_file_versions", maxFileVersions)
		}
		if autrotateVersions != nil {
			d.Set("remove_older_versions", *autrotateVersions)
		}
		if defaultTtlForFiles != nil {
			d.Set("default_ttl_for_files", *defaultTtlForFiles)
		}
	} else {
		return diag.Errorf("Folder doesn't exists. Read Folder failed!!!")
	}
	return nil
}

func resourceAccessPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	client := m.(*client.Client)
	if client == nil {
		return diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	stat, err := getMetaFrom(name, client)
	if err != nil {
		return err
	}

	if stat.Status == 0 {
		removeFolder := request.RemoveFolder{
			FolderFullPath: "/" + name,
			IsPerm:         false,
		}
		req := request.Request{
			Operation: &request.Request_RemoveFolder{
				RemoveFolder: &removeFolder,
			},
		}
		resp, err := client.Send(&req)
		if err != nil {
			return diag.Errorf("Request/Response sent/recieved incorrectly" + err.Error())
		}
		if resp == nil {
			return diag.Errorf("Remove Folder failed!!!")
		}
		if resp.GetRemoveFolder().Status != 0 {
			return diag.Errorf("Remove Folder failed!!!")
		}

	} else {
		return diag.Errorf("Folder doesn't exist. Destroy not valid!!!")
	}
	return nil
}

func resourceAccessPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	client := m.(*client.Client)
	if client == nil {
		return diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	stat, err := getMetaFrom(name, client)
	if err != nil {
		return err
	}
	if stat.Status != 0 {
		return diag.Errorf("Folder doesn't exist. Update not valid!!!")
	} else {
		keyValMappings := []*request.KeyValMapping{}
		if d.HasChange("max_size") {
			if val, ok := d.GetOk("max_size"); ok {
				intVal, ok := val.(int)
				if !ok {
					return diag.Errorf("unexpected type for max_size: %T (want int)", val)
				}
				uintVal := uint64(intVal)
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uintVal)
				keyValMappings = append(keyValMappings, &request.KeyValMapping{
					Attribute: "max_size",
					Value:     buf,
				})
			}
		}
		if d.HasChange("max_file_size") {
			if val, ok := d.GetOk("max_file_size"); ok {
				intVal, ok := val.(int)
				if !ok {
					return diag.Errorf("unexpected type for max_size: %T (want int)", val)
				}
				uintVal := uint64(intVal)
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uintVal)
				keyValMappings = append(keyValMappings, &request.KeyValMapping{
					Attribute: "max_file_size",
					Value:     buf,
				})
			}
		}
		if d.HasChange("max_file_versions") {
			if val, ok := d.GetOk("max_file_versions"); ok {
				intVal, ok := val.(int)
				if !ok {
					return diag.Errorf("unexpected type for max_size: %T (want int)", val)
				}
				uintVal := uint64(intVal)
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uintVal)
				keyValMappings = append(keyValMappings, &request.KeyValMapping{
					Attribute: "max_file_versions",
					Value:     buf,
				})
			}
		}
		if d.HasChange("remove_older_versions") {
			if v, ok := d.GetOk("remove_older_versions"); ok {
				buf := []byte{0}
				if v.(bool) {
					buf[0] = 1
				}
				keyValMappings = append(keyValMappings, &request.KeyValMapping{
					Attribute: "remove_older_versions",
					Value:     buf,
				})
			}
		}
		if d.HasChange("remove_older_versions") {
			if val, ok := d.GetOk("default_ttl_for_files"); ok {
				intVal, ok := val.(int)
				if !ok {
					return diag.Errorf("unexpected type for max_size: %T (want int)", val)
				}
				uintVal := uint64(intVal)
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uintVal)
				keyValMappings = append(keyValMappings, &request.KeyValMapping{
					Attribute: "default_ttl_for_files",
					Value:     buf,
				})
			}
		}
		_ = request.Policy{
			AttrToValue: keyValMappings,
		}
	}
	d.SetId(name)
	return nil
}
