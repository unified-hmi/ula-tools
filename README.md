# Unified Layout Tools (ULA)

> Unified Layout Tools (ULA) is a framework for achieving flexible layout of applications.
> This framework enables mapping of physical displays onto a virtual screen.
> It provides the ability to apply layout settings such as display position and size on the virtual screen to launched applications.

## Contents

- [Unified Layout Tools (ULA)](#architecture)
  - [Contents](#contents)
  - [Repository structure](#repository-structure)
  - [How to install](#how-to-install)
    - [Golang setup](#golang-setup)
    - [Build ULA framework](#build-ula-framework)
  - [How to Use](#how-to-use)
    - [Json settings](#json-settings)
    - [Workers side](#workers-side)
    - [Manager side](#manager-side)
    - [Command request](#command-request)
    - [How to control layouts on Weston ivi-shell](#how-to-control-layouts-on-weston-ivi-shell)
      - [How to install uhmi-ivi-wm](#how-to-install-uhmi-ivi-wm)
      - [Run Weston](#run-weston)
      - [Run Wayland application on Weston](#run-wayland-application-on-weston)
      - [Control layouts on Weston ivi-shell by ULA](#control-layouts-on-weston-ivi-shell-by-ula)
        - [Run uhmi-ivi-wm](#run-uhmi-ivi-wm)
        - [Json settings (weston)](#json-settings-weston)
        - [Workers side (weston)](#workers-side-weston)
        - [Manager side (weston)](#manager-side-weston)
        - [Command request (weston)](#command-request-weston)
    - [How to control layouts on RVGPU compositor](#how-to-control-layouts-on-rvgpu-compositor)
      - [How to install RVGPU compositor](#how-to-install-rvgpu-compositor)
      - [Run RVGPU compositor](#run-rvgpu-compositor)
      - [Run Wayland application on RVGPU compositor](#run-wayland-application-on-rvgpu-compositor)
      - [Control layouts on RVGPU compositor by ULA](#control-layouts-on-rvgpu-compositor-by-ula)
        - [Json settings (RVGPU)](#json-settings-rvgpu)
        - [Workers side (RVGPU)](#workers-side-rvgpu)
        - [Manager side (RVGPU)](#manager-side-rvgpu)
        - [Command request (RVGPU)](#command-request-rvgpu)

## Repository structure

```
.
├── cmd
│   ├── Makefile
│   ├── ula-client-manager
│   │   ├── Makefile
│   │   └── ula_client_manager.go
│   ├── ula-grpc-client
│   │   ├── Makefile
│   │   └── ula_grpc_client.go
│   └── ula-node
│       ├── main.go
│       └── Makefile
├── CONTRIBUTING.md
├── example
│   ├── dwm
│   │   └── EGLWLInputEventExample
│   │       └── dwm_initial_layout.json
│   ├── initial-vscreen
│   │   ├── global
│   │   │   └── initial-vscreen.json
│   │   └── vdisplay
│   │       └── initial-vscreen.json
│   └── vsd
│       ├── iviwinmgr
│       │   └── virtual-screen-def.json
│       └── rvgpuwinmgr
│           └── virtual-screen-def.json
├── internal
│   ├── Makefile
│   ├── ula
│   │   ├── common_type.go
│   │   ├── env.go
│   │   ├── ula_client_server_protocol.go
│   │   ├── ula_layout_data_creater.go
│   │   └── virtual_screen_def.go
│   ├── ula-client
│   │   ├── core
│   │   │   └── core_type.go
│   │   ├── dwmapi
│   │   │   ├── dwm_client_api.go
│   │   │   ├── dwm_common.go
│   │   │   ├── dwm_server_api.go
│   │   │   └── Makefile
│   │   ├── Makefile
│   │   ├── readclusterapp
│   │   │   ├── Makefile
│   │   │   └── read_cluster_app.go
│   │   ├── ulacommgen
│   │   │   ├── Makefile
│   │   │   ├── ula_comm_generator.go
│   │   │   └── ula_comm_generator_protocol.go
│   │   ├── ulamulticonn
│   │   │   ├── Makefile
│   │   │   └── ula_multi_conn.go
│   │   └── ulavscreen
│   │       ├── Makefile
│   │       ├── ula_vscreen.go
│   │       └── vscreen_to_rdisplay_converter.go
│   ├── ula-node
│   │   ├── common_type.go
│   │   ├── iviwinmgr
│   │   │   ├── ivi_command_generator.go
│   │   │   ├── ivi_layer_split.go
│   │   │   ├── iviwinmgr.go
│   │   │   ├── iviwinmgr_protocol.go
│   │   │   └── Makefile
│   │   ├── Makefile
│   │   ├── rvgpuwinmgr
│   │   │   ├── Makefile
│   │   │   ├── rvgpu_command_generator.go
│   │   │   ├── rvgpuwinmgr.go
│   │   │   └── rvgpuwinmgr_protocol.go
│   │   └── ula_command_data_creater.go
│   └── ulog
│       ├── Makefile
│       └── ulog.go
├── LICENSE.md
├── Makefile
├── pkg
│   └── ula-client-lib
│       ├── Makefile
│       └── ula_client.go
├── proto
│   ├── dwm.proto
│   └── grpc
│       └── dwm
│           ├── dwm_grpc.pb.go
│           └── dwm.pb.go
└── README.md
```

# How to install
## Golang setup
Before building ULA API, you need to install and configure Golang.
```
sudo apt-get install golang
export GOPATH=<your go work directory>
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
```

## Build ULA framework
You can easily build ULA framework using make.
```
mkdir -p $GOPATH/src
cd $GOPATH/src
git clone https://github.com/unified-hmi/ula-tools.git
cd ula-tools
make
```

# How to use
With ULA, you can flexibly configure and apply layouts for applications.
ULA API is implemented by using gRPC protocol. So, you can implement ULA Client APIs within your applications by importing gRPC protocol header files.
When a ULA Client API is called, the gRPC server (ula-client-manager) controls multiple workers(ula-node) and applies controlling layout commands referring to some Json files.

## <a name="json-settings"></a>Json settings
To run ULA, some Json files are required.
* virtual-screen-def.json: execution environment such as display and node (SoCs/VMs/PCs) informations.
* dwm_initial_layout.json: application layout information such as position and size.

Json files must be created correctly for your execution environment.
Sample Json files are located in the "$GOPATH/src/ula-tools/example" directory.

To control layouts, dwm_initial_layout.json have following parameters:

- vlayer: define a virtual layer that represents a group of surfaces within the virtual screen. Each layer has a unique Virtual ID (VID) and can contain multiple surfaces. virtual_w and virtual_h define vlayer's size. The layer's source (vsrc_x, vsrc_y, vsrc_w, vsrc_h) and destination (vdst_x, vdst_y, vdst_w, vdst_h) coordinates determine where and how large the layer appears on the virtual screen.
- vsurface: define individual surfaces within the virtual layer. Each surface also has a VID, and its pixel dimensions (pixel_w, pixel_h) represent the actual size of the content. The source (psrc_x, psrc_y, psrc_w, psrc_h) and destination (vdst_x, vdst_y, vdst_w, vdst_h) coordinates determine the portion of the content to display and its location within the layer.
- coord: vlayer is possible to set the position in two coordinate systems. In the global coordinate system, it defines where it is in relation to the origin of the virtual screen. In the vdisplay coordinate system, it defines where it is in relation to the origin of the display with the specified ID (vdisplay_id).
- visibility: define visibility of vlayer/vsurface. Display when it is 1, and hidden when it is 0.
- z_order: define the display order of the app. The app with a higher z_order value will be displayed primarily on the monitor.

**Note:** [Here](https://docs.automotivelinux.org/en/master/#06_Component_Documentation/11_Unified_HMI/) is the documentation for verifying the operation of the Unified HMI framework on AGL and the detailed explanation about Json files.

## <a name="workers-side"></a>Workers side
Before running Command request, the worker side needs to launch __*ula-node*__.
ula-node has a porting layer to determine which plugin to use, `iviwinmgr` or `rvgpuwinmgr`.
If you define the "compositor" section in virtual-screen-def.json, ula-node branches to use `rvgpuwinmgr` plugin, and if you don't define it, `iviwinmgr` plugin is used.

ula-node receives initial display layout commands and generates local commands from virtual-screen-def.json to send controlling layout commands to the `uhmi-ivi-wm` or `rvgpu-renderer`.

This ula-node needs virtual-screen-def.json, so please set it with "-f" option. (default path is /etc/uhmi-framework/virtual-screen-def.json)

- Options of ula-node
  - -H: search ula-node param by hostname from VScrnDef file
  - -N: search ula-node param by node_id from VScrnDef file (default: -1)
  - -d: verbose debug log
  - -f: virtual-screen-def.json file Path (default: "/etc/uhmi-framework/virtual-screen-def.json")
  - -v: verbose info log (default true)

```
ula-node -f <path to virtual-screen-def.json> &
```

**Note:** Master node may also work as worker.


## <a name="manager-side-1"></a>Manager side
Before running Command request, the manager side needs to launch __*ula-client-manager*__.
ula-client-manager works as a gRPC Server and get Json files from gRPC Client APIs.
It sends layout commands to multiple workers referring to some Json files.

ula-client-manager needs the path to virtual-screen-def.json, so please set it with "-f" option. (default path is /etc/uhmi-framework/virtual-screen-def.json)

- Options of ula-client-manager
  - -d: verbose debug log
  - -f: virtual-screen-def.json file Path (default "/etc/uhmi-framework/virtual-screen-def.json")
  - -v: verbose info log (default true)

```
ula-client-manager -f <path to virtual-screen-def.json>
```


## <a name="command-request-1"></a>Command request
After launching manager and all workers.
__*ula-grpc-client*__ can send controlling layout commands to the manager as gRPC Client API.

- Options of ula-grpc-client
  - -c: specify a dwm api command (default: DwmSetSystemLayout)
       `DwmSetSystemLayout`
       `DwmSetLayoutCommand          <filePath>`
  - -h: Show this message

```
ula-grpc-client -c <command>
```

**Note:** ula-grpc-client is reference implementation of Go language for gRPC Client API and you can implement with various languages which supporting gRPC protocol.
**Note:** `DwmSetLayoutCommand` command needs file path to initial_vscreen.json (not to dwm_initial_vscreen.json). Sample initial_vscreen.json files are located in the "$GOPATH/src/ula-tools/example/initial_vscreen" directory.

ULA also provides a C language shared library (default: generated in $GOPATH/pkg/libulaclient).
By using the library's API, it's easy to implement ULA gRPC Client APIs in your applications.

- Description of Client API
  - dwm_set_system_layout: display the layout according to dwm_initial_layout.json in DWMPATH
  - dwm_set_layout_command: display the layout according to initial_vscreen.json given as an argument

Please create a C language source file with content similar to the following example:
```c
#include <stdio.h>
#include "libulaclient.h"

int main(void)  
{
	dwm_set_system_layout();
	return 0;
}
```

Compile this source file with gcc as follows:
```
gcc -I./ -L./ sample.c -lulaclient -o sample.out
```

Execute the command as follows:
```
export VSDPATH="<path to virtual-screen-def.json>"
export LD_LIBRARY_PATH="<path to libulaclient>"
./sample.out
```

## How to control layouts on Weston ivi-shell
ULA has a plugin (`iviwinmgr`) for supporting Weston ivi-shell.
When you want to use ULA to control layouts on Weston, you should prepare `uhmi-ivi-wm`, which is one the Unified HMI frameworks.
The ULA plugin can change layouts through `uhmi-ivi-wm` on Weston ivi-shell.

### How to install uhmi-ivi-wm
For instructions on how to install `uhmi-ivi-wm`, please refer to the [README](https://github.com/unified-hmi/uhmi-ivi-wm).

### Run Weston
```
weston --width=1920 --height=1080 --output-count=2 &
```
**Note:** The option "--output-count" indicates the number of Westons to be launched. Please specify the number according to the configuration file.

### Run Wayland application on Weston
You can run the wayland-ivi-extension sample app EGLWLInputEventExample to see how the layout works. Of course it will work with other wayland apps that support ivi_application.
```
EGLWLInputEventExample &
```
**Note:** The surface ID of the launched application is used in the json file.

### Control layouts on Weston ivi-shell by ULA
After launching wayland applications on Weston, you can control their layouts by using ULA tools and `uhmi-ivi-wm`.

#### Run uhmi-ivi-wm
To send the application layout information generated by ULA, `uhmi-ivi-wm` needs to be launched.

```
uhmi-ivi-wm &
```

#### <a name="json-settings-weston"></a>Json settings
Sample Json files are located in the "$GOPATH/src/ula-tools/example" directory.
Please modify Json files according to your own execution environment referring to samples.
To use 'iviwinmgr' plugin, please don't define "compositor" section in virtual-screen-def.json.
The sample virtual-screen-def.json is located in the "$GOPATH/src/ula-tools/example/vsd/iviwinmgr/virtual-screen-def.json".

#### <a name="workers-side-weston"></a>Workers side
Execute the ULA command for Worker on the all hosts on which you want to control application layouts.
```
export GOPATH=<your go work directory>
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
ula-node -f $GOPATH/src/ula-tools/example/vsd/iviwinmgr/virtual-screen-def.json
```

#### <a name="manager-side-weston"></a>Manager side
After all Worker commands have been executed, launch Manager command on any host.
```
export GOPATH=<your go work directory>
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
export DWMPATH="<path to the directory that contains the dwm_initial_layout.json>"
ula-client-manager -f $GOPATH/src/ula-tools/example/vsd/iviwinmgr/virtual-screen-def.json
```

#### <a name="command-request-weston"></a>Command request
After all Worker and Manager commands have been executed, launch command request on any host.
```
export GOPATH=<your go work directory>
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
export VSDPATH="<path to virtual-screen-def.json>"
ula-grpc-client
```

## How to control layouts on RVGPU compositor
Remote VIRTIO GPU (RVGPU) is one of the main components of Unified HMI frameworks.
RVGPU is a client-server based rendering engine, which allows to render 3D on one device (client) and display it via network on another device (server).
RVGPU has compositor functions, and it can composite multiple applications on it.
By combining it with ULA, you can control the layouts of applications on the RVGPU compositor.

### How to install RVGPU compositor
For instructions on how to install RVGPU, please refer to the [README](https://github.com/unified-hmi/remote-virtio-gpu).

### Run RVGPU compositor
For instructions on how to run RVGPU compositor, please refer to the [README](https://github.com/unified-hmi/remote-virtio-gpu).

#### Run rvgpu-renderer
To combine ULA with RVGPU compositor, you have to run `rvgpu-renderer` with "-l" option.
Additionally, you need to add "-d" option to use the same socket domain to connect ula-node with `rvgpu-renderer`.

e.g. launch `rvgpu-renderer` with the socket domain "rvgpu-compositor-0"
```
rvgpu-renderer -b 1920x1080@0,0 -p 36000 -a -d rvgpu-compositor-0 -l
```

#### Run rvgpu-proxy
After launching `rvgpu-renderer`, you can run multiple `rvgpu-proxy` and applications on one `rvgpu-renderer`.
Please add the application name, which is the same as defined in initial-vscreen.json or dwm_initial_layout.json, with "-i" option.

e.g. launch `rvgpu-proxy` with the application name "wayland-app"
```
rvgpu-proxy -s 1920x1080@0,0 -n 127.0.0.1:36000 -i wayland-app
```

### Run Wayland application on RVGPU compositor
If you wish to display Wayland applications remotely, you can optionally install `rvgpu-wlproxy`, which can be used as a lightweight Wayland server instead of Weston for similar functionality. For detailed installation, configuration, and usage instructions, please refer to the [README](https://github.com/unified-hmi/rvgpu-wlproxy).

### Control layouts on RVGPU compositor by ULA
After launching Wayland applications with these steps, you can control layouts on RVGPU compositor by using ULA tools.
For instructions on how to run ULA tools in detail, please refer to the [Control layouts on Weston ivi-shell by ULA](#control-layouts-on-weston-ivi-shell-by-ula)


#### <a name="json-settings-rvgpu"></a>Json settings
Sample Json files are located in the "$GOPATH/src/ula-tools/example" directory.
Please modify Json files according to your own execution environment referring to samples.
To use 'iviwinmgr' plugin, please not to define "compositor" section in virtual-screen-def.json.
The sample virtual-screen-def.json is located in the "$GOPATH/src/ula-tools/example/vsd/rvgpuwinmgr/virtual-screen-def.json".

#### <a name="workers-side-rvgpu"></a>Workers side
To connect with `rvgpu-renderer`, please define the "compositor" section in virtual-screen-def.json and add socket domain name in that section, using the "sock_domain_name" key, with the value added on `rvgpu-renderer` as the "-d" option.

After creating virtual-screen-def.json, please run the command below.
```
export GOPATH=<your go work directory>
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
ula-node -f $GOPATH/src/ula-tools/example/vsd/iviwinmgr/virtual-screen-def.json
```

#### <a name="manager-side-rvgpu"></a>Manager side
After all Worker commands have been executed, launch Manager command on any host.
```
export GOPATH=<your go work directory>
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
export DWMPATH=$GOPATH/src/ula-tools/example/dwm
ula-client-manager -f $GOPATH/src/ula-tools/example/vsd/iviwinmgr/virtual-screen-def.json
```

#### <a name="command-request-rvgpu"></a>Command request
To control layouts of the specific application running on RVGPU compositor, please add the "application_name" key, with the value added on `rvgpu-proxy` as the "-i" option.

After all Worker and Manager commands have been executed, launch command request on any host.
```
export GOPATH=<your go work directory>
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
export VSDPATH="<path to virtual-screen-def.json>"
ula-grpc-client
```
