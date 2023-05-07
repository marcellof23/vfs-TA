package constant

var (
	UsageCommandMkdir = `Usage : mkdir [list of directories to make]`
	UsageCommandPwd   = `Usage : pwd`
	UsageCommandLs    = `Usage : ls
        ls [Directory name] 
	`
	UsageCommandCat   = `Usage : cat [list of directories to make]`
	UsageCommandStat  = `Usage : stat [list of directories to make]`
	UsageCommandTouch = `Usage : upload [list of directories to make]`
	UsageCommandRm    = `Usage : rm [File name] 
        rm -r [Directories Name] remove directories and their contents recursively
	`
	UsageCommandCp = `Usage : cp [File name source] [File name destination]
        cp -r [Directories source] [Directories destination] copy directories and their contents recursively
	`
	UsageCommandMv      = `Usage : mv [list of directories to make]`
	UsageCommandChmod   = `Usage : chmod [list of directories to make]`
	UsageCommandUpload  = `Usage : upload [list of directories to make]`
	UsageCommandMigrate = `Usage : migrate [source cloud provider] [destination cloud provider]
		list of cloud provider : [gcs, dos, s3]
	`
	UsageCommandDownload = `Usage : download [File name local] [File name vfs]
        download -r [Directories local] [Directories vfs] copy directories and their contents recursively
	`
)
