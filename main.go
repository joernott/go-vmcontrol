// vmcontrol project main.go
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/davecgh/go-spew/spew"

	"github.com/docopt/docopt-go"
	"github.com/joernott/go-proxmox"

	"github.com/joernott/go-foreman"
)

var proxmoxhost string
var proxmoxuser string
var proxmoxpass string
var foremanhost string
var foremanuser string
var foremanpass string

var p *proxmox.ProxMox
var f *foreman.Foreman

func getLoginParams(arguments map[string]interface{}) (string, string, string, string, string, string, error) {
	var proxmoxhost string
	var proxmoxuser string
	var proxmoxpass string
	var foremanhost string
	var foremanuser string
	var foremanpass string
	var ok bool
	if proxmoxhost, ok = arguments["--proxmoxhost"].(string); !ok {
		return "", "", "", "", "", "", errors.New("No proxmox host provided.")
	}
	if proxmoxuser, ok = arguments["--proxmoxuser"].(string); !ok {
		return "", "", "", "", "", "", errors.New("No proxmox user provided.")
	}
	if proxmoxpass, ok = arguments["--proxmoxpass"].(string); !ok {
		return "", "", "", "", "", "", errors.New("No proxmox password provided.")
	}
	if foremanhost, ok = arguments["--foremanhost"].(string); !ok {
		return proxmoxhost, proxmoxuser, proxmoxpass, "", "", "", nil //no error if no foreman host is there, we don't need the others
	}
	if foremanuser, ok = arguments["--foremanuser"].(string); !ok {
		return proxmoxhost, proxmoxuser, proxmoxpass, "", "", "", errors.New("No foreman user provided.")
	}
	if foremanpass, ok = arguments["--foremanpass"].(string); !ok {
		return proxmoxhost, proxmoxuser, proxmoxpass, "", "", "", errors.New("No foreman password provided.")
	}
	return proxmoxhost, proxmoxuser, proxmoxpass, foremanhost, foremanuser, foremanpass, nil
}

func getCreateVMParameters(arguments map[string]interface{}) (string, int64, int64, int64, string, int, bool, error) {
	var s string
	var name string
	var cpu int64
	var cores int64
	var mem int64
	var disksize string
	var hostgroup int
	var start bool
	var ok bool
	var err error

	if name, ok = arguments["--name"].(string); !ok {
		return "", 0, 0, 0, "", 0, false, errors.New("No Name.")
	}
	if s, ok = arguments["--cpu"].(string); !ok {
		return "", 0, 0, 0, "", 0, false, errors.New("No CPUs.")
	}
	cpu, err = strconv.ParseInt(s, 10, 32)
	if err != nil {
		return "", 0, 0, 0, "", 0, false, errors.New("Illegal cpu count.")
	}
	if cpu < 1 {
		return "", 0, 0, 0, "", 0, false, errors.New("Not enough CPUs.")
	}

	if s, ok = arguments["--cores"].(string); !ok {
		return "", 0, 0, 0, "", 0, false, errors.New("No Cores.")
	}
	cores, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return "", 0, 0, 0, "", 0, false, errors.New("Illegal core count.")
	}
	if cores < 1 {
		return "", 0, 0, 0, "", 0, false, errors.New("Not enough cores.")
	}

	if s, ok = arguments["--mem"].(string); !ok {
		return "", 0, 0, 0, "", 0, false, errors.New("No memory.")
	}
	mem, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return "", 0, 0, 0, "", 0, false, errors.New("Illegal memory.")
	}
	if mem < 64 {
		return "", 0, 0, 0, "", 0, false, errors.New("Everybody will need more than 64 MiB of memory.")
	}

	if disksize, ok = arguments["--disk"].(string); !ok {
		return "", 0, 0, 0, "", 0, false, errors.New("No disk.")
	}
	if s, ok = arguments["--hostgroup"].(string); !ok {
		return "", 0, 0, 0, "", 0, false, errors.New("No Hostgroup ID.")
	}
	hostgroup, err = strconv.Atoi(s)
	if err != nil {
		return "", 0, 0, 0, "", 0, false, errors.New("Illegal host group.")
	}
	if _, ok = arguments["--start"].(bool); ok {
		start = true
	}
	return name, cpu, cores, mem, disksize, hostgroup, start, nil
}

func CreateVM(arguments map[string]interface{}) {
	var name string
	var cpu int64
	var cores int64
	var mem int64
	var disksize string
	var hostgroup int
	var start bool
	var err error
	var node proxmox.Node
	var qemuList proxmox.QemuList
	var qemu proxmox.QemuVM
	var vmId string
	var qemuConfig proxmox.QemuConfig
	var mac string
	var foremanId string
	var data map[string]interface{}
	var interfaces []interface{}
	var IP string
	var ok bool

	name, cpu, cores, mem, disksize, hostgroup, start, err = getCreateVMParameters(arguments)
	if err != nil {
		fmt.Println(err)
		os.Exit(11)
	}

	node, err = p.DetermineVMPlacement(cpu, cores, mem, 0, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(12)
	}
	vmId, err = node.CreateQemuVM(name, int(cpu), int(cores), int(mem), disksize)
	if err != nil {
		fmt.Println("Could not create VM on " + node.Node + ".")
		fmt.Println(err)
		os.Exit(13)
	}
	qemuList, err = node.Qemu()
	if err != nil {
		fmt.Println(err)
		os.Exit(13)
	}
	qemu = qemuList[vmId]
	qemuConfig, err = qemu.Config()
	mac = qemuConfig.Net["net0"]["virtio"]
	foremanId, err = f.CreateHost(hostgroup, name, mac)
	if err != nil {
		fmt.Println("Could not set up host in foreman")
		fmt.Println(err)
		_, err := qemu.Delete()
		if err != nil {
			fmt.Println("Could not delete VM in Proxmox: Manual cleanup needed!")
			fmt.Println(err)
		}
		os.Exit(14)
	}
	if start {
		err = qemu.Start()
		if err != nil {
			fmt.Println("Error starting VM")
			fmt.Println(err)
			err = f.DeleteHost(vmId)
			if err != nil {
				fmt.Println("Could not delete host in foreman: Manual cleanup needed!")
			}
			_, err = qemu.Delete()
			if err != nil {
				fmt.Println("Could not delete VM in Proxmox: Manual cleanup needed!")
				fmt.Println(err)
			}
			os.Exit(15)
		}
	}
	data, err = f.Get("hosts/" + foremanId)
	if err != nil {
		fmt.Println(err)
		IP = ""
	}
	interfaces = data["interfaces"].([]interface{})
	if data, ok = interfaces[0].(map[string]interface{}); ok {
		IP = data["ip"].(string)
	} else {
		fmt.Println("Did not find interface 0")
	}
	fmt.Println("VMID=" + vmId)
	fmt.Println("MAC=" + mac)
	fmt.Println("IP=" + IP)
	fmt.Println("FOREMANID=" + foremanId)
	os.Exit(0)
}

func DeleteVM(arguments map[string]interface{}) {
	var vmId string
	var foremanId string
	var err error
	var ok bool
	var qemu proxmox.QemuVM

	if vmId, ok = arguments["--vmid"].(string); !ok {
		fmt.Println("No Vm ID")
		os.Exit(21)
	}
	if foremanId, ok = arguments["--foremanid"].(string); !ok {
		fmt.Println("No Foreman ID")
		os.Exit(21)
	}

	ok = true
	err = f.DeleteHost(foremanId)
	if err != nil {
		fmt.Println(err)
		ok = false
	}
	qemu, err = p.FindVM(vmId)
	if err != nil {
		fmt.Println(err)
		ok = false
	}
	if ok {
		err = doVM(p, vmId, "stop")
		if err != nil {
			fmt.Println(err)
			os.Exit(23)
		}
		err = qemu.WaitForStatus("stopped", 60)
		if err != nil {
			fmt.Println(err)
			os.Exit(23)
		}
		_, err = qemu.Delete()
		if err != nil {
			fmt.Println(err)
			os.Exit(23)
		}
		os.Exit(0)
	} else {
		os.Exit(22)
	}
}

func doVM(p *proxmox.ProxMox, vmId string, action string) error {
	var qemu proxmox.QemuVM
	var err error

	qemu, err = p.FindVM(vmId)
	if err != nil {
		return err
	}
	switch action {
	case "start":
		err = qemu.Start()
		if err != nil {
			err = qemu.WaitForStatus("started", 30)
		}
	case "stop":
		err = qemu.Stop()
		if err != nil {
			err = qemu.WaitForStatus("stopped", 30)
		}
	}
	return err
}

func StartVM(arguments map[string]interface{}) {
	var vmId string
	var ok bool
	var err error
	if vmId, ok = arguments["--vmid"].(string); !ok {
		fmt.Println("No Vm ID")
		os.Exit(31)
	}
	err = doVM(p, vmId, "start")
	if err != nil {
		fmt.Println(err)
		os.Exit(32)
	}
}

func StopVM(arguments map[string]interface{}) {
	var vmId string
	var ok bool
	var err error
	if vmId, ok = arguments["--vmid"].(string); !ok {
		fmt.Println("No Vm ID")
		os.Exit(41)
	}
	err = doVM(p, vmId, "start")
	if err != nil {
		fmt.Println(err)
		os.Exit(42)
	}
}

func CloneVM(arguments map[string]interface{}) {
	os.Exit(0)
}

func DumpVM(arguments map[string]interface{}) {
	var vmId string
	var ok bool
	var err error
	var vm proxmox.QemuVM
	var upid string
	var tasks proxmox.TaskList
	var task proxmox.Task
	var exitstatus string

	if vmId, ok = arguments["--vmid"].(string); !ok {
		fmt.Println("No Vm ID")
		os.Exit(51)
	}
	vm, err = p.FindVM(vmId)
	if err != nil {
		fmt.Println(err)
		os.Exit(42)
	}
	upid, err = vm.Node.VZDump(vmId, 65536, "lzo", 0, 2, "stop")
	if err != nil {
		fmt.Println(err)
		os.Exit(43)
	}
	time.Sleep(time.Second * 4)
	tasks, err = p.Tasks()
	if err != nil {
		fmt.Println(err)
		os.Exit(44)
	}
	if task, ok = tasks[upid]; !ok {
		fmt.Println("Could not get task for upid " + upid)
		os.Exit(44)
	}
	exitstatus, err = task.WaitForStatus("stopped", 1800)
	if err != nil {
		fmt.Println(err)
		os.Exit(45)
	}
	if exitstatus != "OK" {
		fmt.Println("Dumping the VM failed with exit status " + exitstatus)
		os.Exit(46)
	}
	os.Exit(0)
}

func main() {
	var err error
	var action interface{}
	var ok bool

	arguments, err := docopt.Parse(usage, nil, true, "vmcontrol v 0.1", false)
	if err != nil {
		fmt.Println(err)
		fmt.Println(usage)
		os.Exit(2)
	}
	//spew.Dump(arguments)
	proxmoxhost, proxmoxuser, proxmoxpass, foremanhost, foremanuser, foremanpass, err = getLoginParams(arguments)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	p, err = proxmox.NewProxMox(proxmoxhost, proxmoxuser, proxmoxpass)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	f = foreman.NewForeman(foremanhost, foremanuser, foremanpass)

	if action, ok = arguments["create-vm"]; ok {
		if action.(bool) {
			CreateVM(arguments)
		}
	}
	if action, ok := arguments["delete-vm"]; ok {
		if action.(bool) {
			DeleteVM(arguments)
		}
	}
	if action, ok := arguments["start-vm"]; ok {
		if action.(bool) {
			StartVM(arguments)
			os.Exit(0)
		}
	}
	if action, ok := arguments["stop-vm"]; ok {
		if action.(bool) {
			StopVM(arguments)
			os.Exit(0)
		}
	}
	if action, ok := arguments["clone-vm"]; ok {
		if action.(bool) {
			CloneVM(arguments)
			os.Exit(0)
		}
	}
	if action, ok := arguments["dump-vm"]; ok {
		if action.(bool) {
			DumpVM(arguments)
			os.Exit(0)
		}
	}
	fmt.Println("No valid action")
	os.Exit(100)
}
