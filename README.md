# go-vmcontrol

vmcontrol is a tool to create and provision virtual machines using the api of ProxMox VE and The Foreman.

I use it mainly in build scripts to facilitate creating virtual machines. Currently, we have fonctions for 
* creating Qemu/KVM VMs on the first ProxMox node in the cluster with sufficient capacity
* starting, stopping, deleting Qemu/KVM VMs
* creating backups (dump) Qemu/KVM VMs


This is currently work in progress. Use it at your own risk. 


## Usage:
>  vmcontrol -h | --help
>  vmcontrol create-vm [--vmid=<vmid>] --name=<vmname> --cpu=<cpucount> --cores=<corecount> --mem=<memory> --disk=<disksize> --hostgroup=<id> [--start] --proxmoxhost=<hostname> --proxmoxuser=<username> --proxmoxpass=<secret> --foremanhost=<hostname> --foremanuser=<username> --foremanpass=<secret>
>  vmcontrol delete-vm --vmid=<vmid> --foremanid=<foremanid> --proxmoxhost=<hostname> --proxmoxuser=<username> --proxmoxpass=<secret> --foremanhost=<hostname> --foremanuser=<username> --foremanpass=<secret>
>  vmcontrol start-vm --vmid=<vmid> --proxmoxhost=<hostname> --proxmoxuser=<username> --proxmoxpass=<secret>
>  vmcontrol stop-vm  --vmid=<vmid> --proxmoxhost=<hostname> --proxmoxuser=<username> --proxmoxpass=<secret>
>  vmcontrol clone-vm --vmid=<vmid> [--newvmid=<vmid>] --name=<vmname> --cpu=<cpucount> --cores=<corecount> --mem=<memory> --disk=<disksize> --hostgroup=<id> [--start] --proxmoxhost=<hostname> --proxmoxuser=<username> --proxmoxpass=<secret> --foremanhost=<hostname> --foremanuser=<username> --foremanpass=<secret>
>  vmcontrol dump-vm --vmid=<vmid> --proxmoxhost=<hostname> --proxmoxuser=<username> --proxmoxpass=<secret>

## Options:
>  -h, --help                show this page
>  --name=<vmname>           hostname for the vm
>  --cpu=<cpucount>          number of CPUs
>  --cores=<corecount>       number of cores per CPU
>  --mem=<memory>            amount of memory in MiB
>  --disk=<disksize>         size of the virtual disk (add G for GiB, M for MiB)
>  --start                   automatically start the VM
>  --proxmoxhost=<hostname>  hostname of a proxmox server
>  --proxmoxuser=<username>  user for proxmox
>  --proxmoxpass=<secret>    password for the proxmox user
>  --foremanhost=<hostname>  hostname of the foreman server
>  --foremanuser=<username>  user for foreman
>  --foremanpass=<secret>    password for the foreman user
>  --hostgroup=<id>          id of the foreman host group for the server
>  --vmid=<vmid>             id of the VM in ProxMox
>  --foremanid=<foremanid>   id of the VM in Foreman
`
