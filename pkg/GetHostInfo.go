package main

import (
	"fmt"
	"io/ioutil"
	osnet "net"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type os_info_type struct {
	BootTime            uint64 `json:"boot_time" label:"启动"`
	Network             string `json:"network" lable:"联网"`
	OsType              string `json:"os_type" label:"系统类别"`
	OsName              string `json:"os_name" label:"操作系统"`
	PlatformFamily      string `json:"platform_family" label:"Family"`
	KernelVersion       string `json:"kernel_version" label:"内核版本"`
	PhysicalMachine     string `json:"physical_machine" label:"虚拟化平台"`
	PkgManagementSystem string `json:"pkg_management_system" label:"包管理系统"`
	PkgRepoList         string `json:"pkg_repo_list" label:"源列表信息"`
	GlibcVersion        string `json:"glibc_ver" label:"GLIBC版本"`
}

type user_info_type struct {
	User    string `json:"user" label:"登录用户"`
	Client  string `json:"client" label:"客户端"`
	Started int    `json:"started" label:"登录时间"`
}

type cpu_info_type struct {
	CpuArch        string  `json:"cpu_arch" label:"CPU架构"`
	CpuNum         int     `json:"cpu_num" label:"CPU核心数"`
	CpuModelName   string  `json:"cpu_model_name" label:"CPU型号"`
	CpuUsedPercent float64 `json:"cpu_used_percent" label:"CPU使用率"`
}

type mem_info_type struct {
	MemTotal       uint64  `json:"mem_total" label:"内存总量"`
	MemUsed        uint64  `json:"mem_used" label:"已使用"`
	MemUsedPercent float64 `json:"mem_used_percent" label:"使用率"`
	MemFree        uint64  `json:"mem_free" label:"未使用"`
	MemAvailable   uint64  `json:"mem_available" label:"可用"`
	SwapTotal      uint64  `json:"swap_total" label:"Swap总量"`
	SwapUsed       uint64  `json:"swap_used" label:"已使用"`
}

type disk_info_type struct {
	Device     string `json:"device" label:"磁盘"`
	MountPoint string `json:"mountpoint" label:"挂载点"`
	Fstype     string `json:"fstype" label:"文件系统"`
	Total      uint64 `json:"total" label:"磁盘大小"`
	Used       uint64 `json:"used" label:"已使用"`
	Free       uint64 `json:"free" label:"可用"`
}

type net_info_type struct {
	Name  string   `json:"name" label:"网卡名称"`
	Addrs []string `json:"addrs" label:"IP"`
}

type port_info_type struct {
	Port    uint32 `json:"port" label:"端口"`
	IP      string `json:"ip" label:"监听地址"`
	Pid     int32  `json:"pid" label:"Pid"`
	Command string `json:"command" label:"命令"`
	Name    string `json:"name" label:"名称"`
}

type host_info_type struct {
	HostName string           `json:"hostname" label:"主机名"`
	OsInfo   os_info_type     `json:"os_info" label:"系统信息"`
	CpuInfo  cpu_info_type    `json:"cpu_info" label:"CPU信息"`
	MemInfo  mem_info_type    `json:"mem_info" label:"内存信息"`
	NetInfo  []net_info_type  `json:"net_info" label:"网卡信息"`
	UserInfo []user_info_type `json:"user_info" label:"登录用户信息"`
	DiskInfo []disk_info_type `json:"disk_info" label:"磁盘信息"`
	PortInfo []port_info_type `json:"port_info" label:"端口信息"`
}

func get_os_info(host_info *host_info_type) {
	// 是否可联网
	_, err := osnet.DialTimeout("tcp", "114.114.114.114:53", 2*time.Second)
	if err != nil {
		host_info.OsInfo.Network = "不可联网"
	} else {
		host_info.OsInfo.Network = "可联网"
	}

	// 获取glibc版本
	glibc_ver := "未知"
	cmd_info := exec.Command("getconf", "GNU_LIBC_VERSION")
	ver, err := cmd_info.Output()
	if err == nil {
		output_array := strings.Split(string(ver), " ")
		glibc_ver = strings.Trim(output_array[len(output_array)-1], "\n")
	}
	host_info.OsInfo.GlibcVersion = glibc_ver

	// 获取操作系统版本
	var os_name, platform_family string
	filepath := "/etc/os-release"
	context, err := ioutil.ReadFile(filepath)
	if err == nil {
		context_slice := strings.Split(string(context), "\n")
		for _, v := range context_slice {
			if strings.Contains(strings.ToLower(v), "pretty_name") {
				os_name = strings.Trim(strings.Join(strings.Split(v, "=")[1:], "="), "\"")
				//fmt.Printf("inner: %s\n", os_name)
			}
			if strings.Contains(strings.ToLower(v), "id_like") {
				platform_family = strings.Trim(strings.Join(strings.Split(v, "=")[1:], "="), "\"")
				//fmt.Printf("inner: %s\n", platform_family)
			}
		}
	}
	// 获取主机信息
	os_info, err := host.Info()
	if err != nil {
		fmt.Printf("Error: os info get error: %s\n", err)
	} else {
		host_info.HostName = os_info.Hostname
		host_info.OsInfo.BootTime = os_info.BootTime
		host_info.OsInfo.OsType = os_info.OS
		host_info.OsInfo.KernelVersion = os_info.KernelVersion

		if os_name == "" {
			os_name = fmt.Sprintf("%s %s", os_info.Platform, os_info.PlatformVersion)
		}
		if platform_family == "" {
			platform_family = os_info.PlatformFamily
			if platform_family == "" {
				platform_family = "未知"
			}
		}
		host_info.OsInfo.OsName = os_name
		host_info.OsInfo.PlatformFamily = platform_family

		//user_info, _ := host.Users()
		//fmt.Printf("user: %#v\n", user_info)
	}

	// 获取包管理方式
	pkg_management_system := ""
	pkg_management_repolist_command := "无"
	for _, v := range [...]string{"apt", "yum"} {
		cmd_info := exec.Command("which", v)
		_, err := cmd_info.Output()
		if err == nil {
			pkg_management_system = v
			if pkg_management_system == "apt" {
				pkg_management_repolist_command = "apt update"
			} else if pkg_management_system == "yum" {
				pkg_management_repolist_command = "yum clean all && yum repolist"
			}
			break
		}
	}
	host_info.OsInfo.PkgManagementSystem = pkg_management_system
	if pkg_management_system != "" {
		cmd_info = exec.Command("sudo", "bash", "-c", pkg_management_repolist_command)
		put, _ := cmd_info.Output()
		host_info.OsInfo.PkgRepoList = string(put)
	}

	// 判断是物理机还是虚拟机
	physical_machine := "未知"
	/*
		virtual_flag := "virtual"
		for _, v := range [...]string{"dmidecode -s system-product-name", "lshw -class system | grep product"} {
			cmd_info := exec.Command("sudo", "bash", "-c", v)
			output, err := cmd_info.Output()
			if err == nil {
				if strings.Contains(strings.ToLower(string(output)), virtual_flag) {
					physical_machine = "虚拟机"
				} else {
					physical_machine = "物理机"
				}
				break
			}
		}
	*/
	cmd_info = exec.Command("sudo", "systemd-detect-virt")
	output, err := cmd_info.Output()
	output_str := strings.Trim(string(output), "\n")
	if strings.ToLower(output_str) == "none" {
		physical_machine = "物理机"
	} else {
		if err == nil {
			physical_machine = output_str
		}
	}
	if physical_machine == "未知" {
		if os_info.VirtualizationRole == "guest" {
			physical_machine = "虚拟机"
		} else if os_info.VirtualizationRole == "host" {
			physical_machine = "物理机"
		}
	}
	host_info.OsInfo.PhysicalMachine = physical_machine
}

func get_user_info(host_info *host_info_type) {
	user_info, err := host.Users()
	if err != nil {
		fmt.Printf("Error: user info get error: %s\n", err)
	} else {
		//fmt.Printf("user info: %v\n", user_info)
		user_num := len(user_info)
		users := make([]user_info_type, user_num)
		for i, v := range user_info {
			users[i].User = v.User
			users[i].Client = v.Host
			if v.Host == "" {
				users[i].Client = v.Terminal
			}
			users[i].Started = v.Started
		}
		host_info.UserInfo = users
	}
}

func get_cpu_info(host_info *host_info_type) {
	cpu_info, err := cpu.Info()
	if err != nil {
		fmt.Printf("Error: cpu info get error: %s\n", err)
	} else {
		host_info.CpuInfo.CpuNum = len(cpu_info)
		host_info.CpuInfo.CpuModelName = cpu_info[0].ModelName
		cpu_used_percent, err := cpu.Percent(2*time.Second, false)
		if err != nil {
			fmt.Printf("Error: cpu used percent err: %s\n", err)
			host_info.CpuInfo.CpuUsedPercent = -1
		} else {
			host_info.CpuInfo.CpuUsedPercent = cpu_used_percent[0]
		}
	}
	cpu_arch, _ := host.KernelArch()
	host_info.CpuInfo.CpuArch = cpu_arch
}

func get_mem_info(host_info *host_info_type) {
	mem_info, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("Error: mem info get error: %s\n", err)
	} else {
		host_info.MemInfo.MemTotal = mem_info.Total
		host_info.MemInfo.MemUsed = mem_info.Used
		host_info.MemInfo.MemUsedPercent = mem_info.UsedPercent
		host_info.MemInfo.MemFree = mem_info.Free
		host_info.MemInfo.MemAvailable = mem_info.Available
	}
	swap_info, err := mem.SwapMemory()
	if err != nil {
		fmt.Printf("Error: swap info get error: %s\n", err)
	} else {
		host_info.MemInfo.SwapTotal = swap_info.Total
		host_info.MemInfo.SwapUsed = swap_info.Used
	}
}

func get_disk_info(host_info *host_info_type) {
	partition_info, err := disk.Partitions(false)
	if err != nil {
		fmt.Printf("Error: disk partition info get error: %s\n", err)
	} else {
		disk_num := len(partition_info)
		disks := make([]disk_info_type, disk_num)
		for i, v := range partition_info {
			disks[i].Device = v.Device
			disks[i].Fstype = v.Fstype
			disk_info, err := disk.Usage(v.Mountpoint)
			if err != nil {
				fmt.Printf("Error: %s info get error: %s\n", v.Mountpoint, err)
			} else {
				disks[i].MountPoint = disk_info.Path
				disks[i].Total = disk_info.Total
				disks[i].Used = disk_info.Used
				disks[i].Free = disk_info.Free
			}
		}
		// 添加网络挂载存储
		partition_info_all, err := disk.Partitions(true)
		if err != nil {
			fmt.Printf("Error: disk partition info get error: %s\n", err)
		} else {
			var temp_disk disk_info_type
			for _, v := range partition_info_all {
				if strings.Contains(v.Fstype, "fuse.") {
					disk_info, err := disk.Usage(v.Mountpoint)
					if err != nil {
						fmt.Printf("Error: %s info get error: %s\n", v.Mountpoint, err)
					} else {
						if disk_info.Total != 0 {
							temp_disk.Device = v.Device
							temp_disk.Fstype = v.Fstype
							temp_disk.MountPoint = disk_info.Path
							temp_disk.Total = disk_info.Total
							temp_disk.Used = disk_info.Used
							temp_disk.Free = disk_info.Free
							disks = append(disks, temp_disk)
						}
					}
				}
			}
		}
		host_info.DiskInfo = disks
	}
}

func get_net_info(host_info *host_info_type) {
	net_info, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error: net info get error: %s\n", err)
	} else {
		net_num := len(net_info)
		nets := make([]net_info_type, net_num)
		for i, v := range net_info {
			nets[i].Name = v.Name
			for _, ip := range v.Addrs {
				nets[i].Addrs = append(nets[i].Addrs, ip.Addr)
			}
		}
		host_info.NetInfo = nets
	}
}

func get_port_info(host_info *host_info_type) {
	ports_info, err := net.Connections("inet")
	if err != nil {
		fmt.Printf("Error: port info get error: %s\n", err)
	} else {
		var port_info port_info_type
		for _, v := range ports_info {
			if v.Status == "LISTEN" {
				port_info.Pid = v.Pid
				port_info.Port = v.Laddr.Port
				port_info.IP = v.Laddr.IP
				port_exist := 0
				for i, v1 := range host_info.PortInfo {
					if port_info.Port == v1.Port {
						host_info.PortInfo[i].IP = host_info.PortInfo[i].IP + "/" + port_info.IP
						port_exist = 1
						break
					}
				}
				if port_exist == 0 {
					p, _ := process.NewProcess(v.Pid)
					port_info.Command, _ = p.Cmdline()
					port_info.Name, _ = p.Name()
					host_info.PortInfo = append(host_info.PortInfo, port_info)
				}
			}
		}
	}
}

func format_size(size uint64) string {
	var c_number float32 = 1024
	kb := float32(size) / c_number
	if kb >= c_number {
		mb := kb / c_number
		if mb >= c_number {
			gb := mb / c_number
			return fmt.Sprintf("%.2fG", gb)
		} else {
			return fmt.Sprintf("%.2fM", mb)
		}
	} else {
		return fmt.Sprintf("%.2fK", kb)
	}
}

func get_label(s interface{}, field_name string) string {
	field_info, _ := reflect.TypeOf(s).FieldByName(field_name)
	tag_name := field_info.Tag.Get("label")
	return tag_name
}

func format_print(host_info host_info_type) {
	os_info := host_info.OsInfo
	cpu_info := host_info.CpuInfo
	mem_info := host_info.MemInfo
	user_info := host_info.UserInfo
	disk_info := host_info.DiskInfo
	net_info := host_info.NetInfo
	port_info := host_info.PortInfo

	os_output := fmt.Sprintf(
		"%s: %s(%s%s, %s)\n"+
			"  %s: %s\t%s: %s\t%s: %s\n"+
			"  %s: %s\t%s: %s\n"+
			"  %s: %s\t%s: %s\n",
		get_label(host_info, "HostName"), host_info.HostName, time.Unix(int64(os_info.BootTime), 0).Local().Format("2006-01-02 15:04:05"), get_label(os_info, "BootTime"), os_info.Network,
		get_label(os_info, "OsType"), os_info.OsType, get_label(os_info, "PlatformFamily"), os_info.PlatformFamily, get_label(os_info, "OsName"), os_info.OsName,
		get_label(os_info, "KernelVersion"), os_info.KernelVersion, get_label(os_info, "PkgManagementSystem"), os_info.PkgManagementSystem, get_label(os_info, "PhysicalMachine"), os_info.PhysicalMachine,
		get_label(os_info, "GlibcVersion"), os_info.GlibcVersion,
	)

	repo_output := fmt.Sprintf(
		"%s:\n  %s",
		get_label(os_info, "PkgRepoList"), strings.TrimRight(strings.Replace(os_info.PkgRepoList, "\n", "\n  ", -1), " "),
	)

	cpu_output := fmt.Sprintf(
		"  %s: %s\t%s: %s\n"+
			"  %s: %d\t%s: %.2f%%\n",
		get_label(cpu_info, "CpuArch"), cpu_info.CpuArch, get_label(cpu_info, "CpuModelName"), cpu_info.CpuModelName,
		get_label(cpu_info, "CpuNum"), cpu_info.CpuNum, get_label(cpu_info, "CpuUsedPercent"), cpu_info.CpuUsedPercent,
	)

	mem_output := fmt.Sprintf(
		"  %s: %s\t%s: %s(%.2f%%)\t%s: %s\t%s: %s\n"+
			"  %s: %s\t%s: %s\n",
		get_label(mem_info, "MemTotal"), format_size(mem_info.MemTotal), get_label(mem_info, "MemUsed"), format_size(mem_info.MemUsed),
		mem_info.MemUsedPercent, get_label(mem_info, "MemFree"), format_size(mem_info.MemFree), get_label(mem_info, "MemAvailable"), format_size(mem_info.MemAvailable),
		get_label(mem_info, "SwapTotal"), format_size(mem_info.SwapTotal), get_label(mem_info, "SwapUsed"), format_size(mem_info.SwapUsed),
	)

	net_output := fmt.Sprintf("%s:\n", get_label(host_info, "NetInfo"))
	for _, v := range net_info {
		net_output += fmt.Sprintf(
			"  %s: %s\t%s: %s\n",
			get_label(v, "Name"), v.Name, get_label(v, "Addrs"), strings.Join(v.Addrs, ", "),
		)
	}

	user_output := ""
	if len(user_info) != 0 {
		user_output += fmt.Sprintf("%s:\n", get_label(host_info, "UserInfo"))
		for _, v := range user_info {
			user_output += fmt.Sprintf(
				"  %s: %s\t%s: %s\t%s: %s\n",
				get_label(v, "User"), v.User, get_label(v, "Client"), v.Client, get_label(v, "Started"), time.Unix(int64(v.Started), 0).Local().Format("2006-01-02 15:04:05"),
			)
		}
	}

	disk_output := fmt.Sprintf("%s:\n", get_label(host_info, "DiskInfo"))
	for _, v := range disk_info {
		disk_output += fmt.Sprintf(
			"  %s: %s(%s)\t%s :%s\t%s: %s\t%s: %s\t%s: %s\n",
			get_label(v, "MountPoint"), v.MountPoint, v.Device, get_label(v, "Fstype"), v.Fstype, get_label(v, "Total"), format_size(v.Total),
			get_label(v, "Used"), format_size(v.Used), get_label(v, "Free"), format_size(v.Free),
		)
	}

	port_output := fmt.Sprintf("%s:\n", get_label(host_info, "PortInfo"))
	for _, v := range port_info {
		port_output += fmt.Sprintf(
			"  %s: %d(%s)\t%s: %d\t%s: %s\n",
			get_label(v, "Port"), v.Port, v.IP, get_label(v, "Pid"), v.Pid, get_label(v, "Name"), v.Name,
		)
	}

	output := os_output + cpu_output + mem_output + repo_output + net_output + user_output + disk_output + port_output
	fmt.Printf("%s", output)
}

func main() {
	fmt.Println("获取主机信息中, 请稍后...")
	host_info := host_info_type{}

	get_os_info(&host_info)
	get_cpu_info(&host_info)
	get_user_info(&host_info)
	get_mem_info(&host_info)
	get_disk_info(&host_info)
	get_net_info(&host_info)
	get_port_info(&host_info)
	format_print(host_info)

	//host_info_json, _ := json.MarshalIndent(host_info, "", "    ")
	//fmt.Printf("host info: %s\n", string(host_info_json))
}
