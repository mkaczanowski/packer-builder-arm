package step

import (
	"text/template"
)

var diskSchema = `#!/bin/sh
set -xeu
DISK=$1

if [ -z ${DISK} ]; then
    echo "no disk specified"
    exit 1
fi;

# format disk as GPT, GUID Partition Table
sgdisk -Z $DISK
{{range .ImagePartitions}}
sgdisk -n 0:{{.StartSector}}:{{.Size}} -t 0:{{.Type}} -c 0:"{{.Name}}" $DISK{{end}}

# print disk layout
sgdisk -p $DISK

# inform the OS of partition table changes
partprobe $DISK
`

var filesystemSchema = `#!/bin/sh
set -xeu

DISK=$1
LOOPDEVICE=$(losetup -f)

if [ -z ${DISK} ] || [ -z ${LOOPDEVICE} ]; then
    echo "no disk or loopdevice specified"
    exit 1
fi;

# cleanup any previous mappings
losetup -j $DISK -O NAME -n | xargs -I{} sh -c 'losetup -d {}'

# map image
losetup -P $LOOPDEVICE $DISK

# format
{{range $i, $a := .ImagePartitions}}
{{ $count := inc $i }}
mkfs.{{$a.Filesystem}} "${LOOPDEVICE}p{{$count}}"{{end}}

# cleanup
losetup -j $DISK -O NAME -n | xargs -I{} sh -c 'losetup -d {}'
`

var (
	diskSchemaTemplate       *template.Template
	filesystemSchemaTemplate *template.Template
)

func init() {
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}

	diskSchemaTemplate = template.Must(template.New("disk-schema").Funcs(funcMap).Parse(diskSchema))
	filesystemSchemaTemplate = template.Must(template.New("filesystem-schema").Funcs(funcMap).Parse(filesystemSchema))
}
