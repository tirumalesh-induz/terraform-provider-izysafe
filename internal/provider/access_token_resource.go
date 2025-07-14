package provider

import (
	"context"
	"encoding/base64"
	"strconv"
	"terraform-provider-izysafe/internal/client"

	"terraform-provider-izysafe/internal/proto/request"
	"terraform-provider-izysafe/internal/proto/response"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const URL = "wss://files.ysafe.io:5577"

func resourceAccessToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccessTokenCreate,
		ReadContext:   resourceAccessTokenRead,
		DeleteContext: resourceAccessTokenDelete,
		UpdateContext: resourceAccessTokenUpdate,

		Schema: map[string]*schema.Schema{
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name of the pin.",
			},
			"pin": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: ValidPin,
				Description:  "A six digit number.",
			},
			"expiry": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ValidateNumericString,
				Description:  "Number of seconds PIN is valid from the creation time.",
			},
			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A secret that is required with pin to authenticate.",
			},
			"id_sent_to_client": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A secret that is required with pin to authenticate.",
			},
		},
	}
}

func resourceAccessTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)
	if client == nil {
		return diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	email := client.Email
	label := d.Get("label").(string)
	pin := d.Get("pin").(string)
	var ttl string
	var addPinReq request.AddPin
	if v, ok := d.GetOk("expiry"); ok {
		ttl = v.(string)
		val, ok := strconv.ParseUint(ttl, 10, 64)
		if ok != nil {
			return diag.Errorf("Failed to parse expiry value: %v", ttl)
		}
		addPinReq = request.AddPin{
			Email:          email,
			Pin:            pin,
			AllowedOps:     []request.AllowedPinOp{},
			Name:           &label,
			Ttl:            val,
			AllowedObjects: [][]byte{},
		}
	} else {
		addPinReq = request.AddPin{
			Email:          email,
			Pin:            pin,
			AllowedOps:     []request.AllowedPinOp{},
			Name:           &label,
			AllowedObjects: [][]byte{},
		}
	}

	req := &request.Request{
		Operation: &request.Request_AddPin{
			AddPin: &addPinReq,
		},
	}

	var idToClientResp []byte
	resp, err := client.Send(req)
	if err != nil {
		return diag.Errorf("Failed to send request:%v", err.Error())
	}
	var Data []byte
	switch r := resp.Operation.(type) {
	case *response.Response_AddPin:
		addPinResponse := r.AddPin
		if addPinResponse.Status != response.Status_SUCCESS {
			return diag.Errorf("Failed to add pin: %v", *addPinResponse.Message)
		}
		Data = addPinResponse.Data
		idToClientResp = addPinResponse.IdToClient
	default:
		return diag.Errorf("Unknown Operation: %v", r)
	}
	d.SetId(label)
	d.Set("token", base64.StdEncoding.EncodeToString(Data))
	d.Set("id_sent_to_client", base64.StdEncoding.EncodeToString(idToClientResp))
	return nil
}

func resourceAccessTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)
	if client == nil {
		return diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	return nil
}

func resourceAccessTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	client := m.(*client.Client)
	if client == nil {
		return diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	email := client.Email
	idToClientData := d.Get("id_sent_to_client").(string)
	tokenData := d.Get("token").(string)
	data, err := base64.StdEncoding.DecodeString(tokenData)
	if err != nil {
		return diag.Errorf("Failed to Decode the Pin Data:%v", err)
	}
	idSentToClient, err := base64.StdEncoding.DecodeString(idToClientData)
	if err != nil {
		return diag.Errorf("Failed to Delete Pin:%v", err)
	}
	delPinReq := &request.DeletePin{
		Email:          email,
		IdSentToClient: idSentToClient,
		Data:           data,
	}

	req := &request.Request{
		Operation: &request.Request_DeletePin{
			DeletePin: delPinReq,
		},
	}

	resp, err := client.Send(req)
	if err != nil {
		return diag.Errorf("Failed to send request: %v", err.Error())
	}

	switch r := resp.Operation.(type) {
	case *response.Response_DeletePin:
		delPinResp := r.DeletePin
		if delPinResp.Status != response.Status_SUCCESS {
			return diag.Errorf("Failed to delete pin: %v", *delPinResp.Message)
		}
	default:
		return diag.Errorf("Unknown Operation: %v", r)
	}

	d.Set("token", "")
	d.Set("id_sent_to_client", "")
	return nil
}

func resourceAccessTokenUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)
	if client == nil {
		return diag.Errorf("Client is nil, please check the token and pin. Contact support if the issue persists.")
	}
	var uuIds [][]byte
	tokenData := d.Get("token").(string)
	label := d.Get("label").(string)
	data, _ := base64.StdEncoding.DecodeString(tokenData)
	updatePinReq := &request.UpdatePinOps{
		Email:          client.Email,
		AllowedOps:     []request.AllowedPinOp{},
		AllowedObjects: uuIds,
		Data:           data,
		PinName:        label,
	}

	req := &request.Request{
		Operation: &request.Request_UpdatePinOps{
			UpdatePinOps: updatePinReq,
		},
	}

	resp, err := client.Send(req)
	if err != nil {
		return diag.Errorf("Failed to send request: %v", err.Error())
	}

	switch r := resp.Operation.(type) {
	case *response.Response_UpdatePinOps:
		updPinResp := r.UpdatePinOps
		if updPinResp.Status != response.Status_SUCCESS {
			return diag.Errorf("Failed to Update Pin:%v", *updPinResp.Message)
		}
		data = updatePinReq.Data
	default:
		diag.Errorf("Unknown Operation: %v", r)
	}
	d.Set("token", base64.StdEncoding.EncodeToString(data))
	return nil
}
