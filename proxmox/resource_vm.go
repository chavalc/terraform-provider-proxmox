package proxmox

import (
	"log"

	"strconv"

	"time"

	"github.com/chavalc/goproxmox"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVMCreate,
		Read:   resourceVMRead,
		Update: resourceVMUpdate,
		Delete: resourceVMDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"node": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vm_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"template": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"start_after_create": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"clone": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
			"args": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"smbios1": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"start_at_boot": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cores": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ide_device": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"file": {
							Type:     schema.TypeString,
							Required: true,
						},
						"media": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"size": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"network_device": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"model": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bridge": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"firewall": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"link_down": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"macaddr": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"queues": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"rate": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"tag": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"trunks": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"serial_device": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"device": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"virtio_device": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"file": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cache": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"backup": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"iothread": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"size": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"snapshot": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func parseVMConfig(d *schema.ResourceData) (*goproxmox.VMConfig, error) {
	config := new(goproxmox.VMConfig)

	if v, ok := d.GetOk("args"); ok {
		config.Args = goproxmox.String(v.(string))
	}
	if v, ok := d.GetOk("cores"); ok {
		config.Cores = goproxmox.Int(v.(int))
	}
	if v, ok := d.GetOk("memory"); ok {
		config.Memory = goproxmox.Int(v.(int))
	}
	if v, ok := d.GetOk("name"); ok {
		config.Name = goproxmox.String(v.(string))
	}
	if v, ok := d.GetOk("smbios1"); ok {
		config.SMBIOS1 = goproxmox.String(v.(string))
	}
	if v, ok := d.GetOk("start_at_boot"); ok {
		config.StartAtBoot = goproxmox.Bool(v.(bool))
	}
	if v, ok := d.GetOk("ide_devices"); ok {
		devices := v.(*schema.Set)
		for _, element := range devices.List() {
			elem := element.(map[string]interface{})
			number := elem["number"].(int)
			media, err := goproxmox.MediaTypeFromString(elem["media"].(string))
			if err != nil {
				return nil, err
			}
			device := &goproxmox.IDEDevice{
				File:  goproxmox.String(elem["file"].(string)),
				Media: &media,
				//Size:     goproxmox.String(elem["size"].(string)),
			}

			log.Printf("[DEBUG] IDE device %v", device.GetQMOptionValue())
			config.AddIDEDevice(number, device)
		}
	}
	if v, ok := d.GetOk("network_devices"); ok {
		devices := v.(*schema.Set)
		for _, element := range devices.List() {
			elem := element.(map[string]interface{})
			log.Printf("[DEBUG] Network device elem %v", elem)
			cardModel, err := goproxmox.NetworkCardModelFromString(elem["model"].(string))
			if err != nil {
				return nil, err
			}
			number := elem["number"].(int)
			device := &goproxmox.NetworkDevice{
				Model: &cardModel,
			}
			if val, ok := elem["bridge"]; ok {
				device.Bridge = goproxmox.String(val.(string))
			}
			//if val, ok := elem["firewall"]; ok {
			//	device.Firewall = goproxmox.Bool(val.(bool))
			//}
			//if val, ok := elem["link_down"]; ok {
			//	device.LinkDown = goproxmox.Bool(val.(bool))
			//}
			if val, ok := elem["macaddr"]; ok && val != "" {
				device.MacAddr = goproxmox.String(val.(string))
			}
			//if val, ok := elem["queues"]; ok {
			//	device.Queues = goproxmox.Int(val.(int))
			//}
			//if val, ok := elem["rate"]; ok {
			//	device.Rate = goproxmox.Float64(val.(float64))
			//}
			//if val, ok := elem["tag"]; ok {
			//	device.Tag = goproxmox.Int(val.(int))
			//}
			//if val, ok := elem["trunks"]; ok {
			//	device.Trunks = goproxmox.String(val.(string))
			//}
			log.Printf("[DEBUG] Network device %v", device.GetQMOptionValue())
			config.AddNetworkDevice(number, device)
		}
	}
	if v, ok := d.GetOk("serial_devices"); ok {
		devices := v.(*schema.Set)
		for _, element := range devices.List() {
			elem := element.(map[string]interface{})
			number := elem["number"].(int)
			device := &goproxmox.SerialDevice{
				Value: elem["device"].(string),
			}

			log.Printf("[DEBUG] Serial device device %v", device.GetQMOptionValue())
			config.AddSerialDevice(number, device)
		}
	}
	if v, ok := d.GetOk("virtio_devices"); ok {
		devices := v.(*schema.Set)
		for _, element := range devices.List() {
			elem := element.(map[string]interface{})
			number := elem["number"].(int)
			//volumeFormat, err := goproxmox.VolumeFormatFromString(elem["format"].(string))
			//if err != nil {
			//	return err
			//}
			device := &goproxmox.VirtIODevice{
				File: goproxmox.String(elem["file"].(string)),
				//Format: &volumeFormat,
				//Backup:   goproxmox.Bool(elem["backup"].(bool)),
				//IOThread: goproxmox.Bool(elem["iothread"].(bool)),
				Size: goproxmox.String(elem["size"].(string)),
				//Snapshot: goproxmox.Bool(elem["snapshot"].(bool)),
			}

			if val, ok := elem["backup"]; ok && val != "" {
				device.Backup = goproxmox.Bool(val.(bool))
			}

			if val, ok := elem["cache"]; ok && val != "" {
				device.Cache = goproxmox.String(val.(string))
			}

			log.Printf("[DEBUG] VirtIO device %v", device.GetQMOptionValue())
			config.AddVirtIODevice(number, device)
		}
	}

	return config, nil
}

func resourceVMCreate(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("clone"); ok {
		return resourceVMCloneCreate(d, meta)
	}

	client := meta.(*goproxmox.Client)
	node := d.Get("node").(string)
	vmID := d.Get("vm_id").(int)

	config, err := parseVMConfig(d)
	if err != nil {
		return err
	}

	if err := client.VMs.CreateVM(node, vmID, config); err != nil {
		return err
	}

	d.SetId(strconv.Itoa(vmID))
	log.Printf("[INFO] VM with ID %s created", d.Id())

	isTemplate := d.Get("template").(bool)
	if isTemplate {
		if err := client.VMs.CreateVMTemplate(node, vmID, ""); err != nil {
			return err
		}
		log.Printf("[INFO] VM template from ID %s created", d.Id())
	}

	startAfterCreate := d.Get("start_after_create").(bool)
	if startAfterCreate {
		if err := client.VMs.StartVM(node, vmID); err != nil {
			return err
		}
	}

	return resourceVMRead(d, meta)
}

func resourceVMCloneCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goproxmox.Client)
	node := d.Get("node").(string)
	newID := d.Get("vm_id").(int)

	cloneConfig := d.Get("clone").(map[string]interface{})
	sourceID := 0
	if val, ok := cloneConfig["source_id"]; ok {
		v, err := strconv.Atoi(val.(string))
		if err != nil {
			return err
		}
		sourceID = v
	}

	c := &goproxmox.VMCloneConfig{}
	if v, ok := d.GetOk("name"); ok {
		c.Name = goproxmox.String(v.(string))
	}
	if _, ok := cloneConfig["full"]; ok {
		c.Full = goproxmox.Bool(true)
	}

	log.Printf("[DEBUG] Cloning VM %d -> %d", sourceID, newID)
	if err := client.VMs.CloneVM(node, sourceID, newID, c); err != nil {
		return err
	}
	d.SetId(strconv.Itoa(newID))
	log.Printf("[INFO] VM with ID %s cloned", d.Id())
	log.Printf("[DEBUG] Waiting for 5 seconds\n")
	time.Sleep(5 * time.Second)

	updateConfig := new(goproxmox.VMConfig)
	if v, ok := d.GetOk("network_devices"); ok {
		devices := v.(*schema.Set)
		for _, element := range devices.List() {
			elem := element.(map[string]interface{})
			log.Printf("[DEBUG] Network device elem %v", elem)
			cardModel, err := goproxmox.NetworkCardModelFromString(elem["model"].(string))
			if err != nil {
				return err
			}
			number := elem["number"].(int)
			device := &goproxmox.NetworkDevice{
				Model: &cardModel,
			}
			if val, ok := elem["bridge"]; ok {
				device.Bridge = goproxmox.String(val.(string))
			}
			if val, ok := elem["macaddr"]; ok && val != "" {
				device.MacAddr = goproxmox.String(val.(string))
			}
			updateConfig.AddNetworkDevice(number, device)
		}
	}
	if v, ok := d.GetOk("cores"); ok {
		updateConfig.Cores = goproxmox.Int(v.(int))
	}
	if v, ok := d.GetOk("memory"); ok {
		updateConfig.Memory = goproxmox.Int(v.(int))
	}
	if updateConfig.NetworkDevices != nil || updateConfig.Cores != nil || updateConfig.Memory != nil {
		if err := client.VMs.UpdateVM(node, newID, updateConfig, false); err != nil {
			return err
		}
	}

	return resourceVMRead(d, meta)
}

func resourceVMRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goproxmox.Client)
	node := d.Get("node").(string)
	vmID := d.Get("vm_id").(int)

	log.Printf("[DEBUG] Fetching VMConfig for node %s, vmID %d", node, vmID)
	config, err := client.VMs.GetVMConfig(node, vmID)
	if err != nil {
		switch err := err.(type) {
		case *goproxmox.VMDoesNotExistError:
			log.Printf("[WARN] %s", err.Error())
			d.SetId("")
			return nil
		case *goproxmox.NodeDoesNotExistError:
			log.Printf("[WARN] %s", err.Error())
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] VMConfig %v", config)
	if config.Name != nil {
		d.Set("name", config.Name)
	}

	return nil
}

func resourceVMUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goproxmox.Client)
	node := d.Get("node").(string)
	vmID := d.Get("vm_id").(int)

	config, err := parseVMConfig(d)
	if err != nil {
		return err
	}
	//if attr, ok := config.GetName(); ok {
	//	d.Set("name", attr)
	//} else {
	//	config.UnsetName()
	//	d.Set("name", nil)
	//}
	//
	//if d.HasChange("name") {
	//	_, n := d.GetChange("name")
	//	config.SetName(n.(string))
	//}

	if err := client.VMs.UpdateVM(node, vmID, config, false); err != nil {
		return err
	}

	return resourceVMRead(d, meta)
}

func resourceVMDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goproxmox.Client)
	node := d.Get("node").(string)
	vmID := d.Get("vm_id").(int)

	status, err := client.VMs.GetVMCurrentStatus(node, vmID)
	if err != nil {
		return err
	}
	if status.Status == "running" {
		if err := client.VMs.StopVM(node, vmID); err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	}

	if err := client.VMs.DeleteVM(node, vmID); err != nil {
		return err
	}

	return nil
}
