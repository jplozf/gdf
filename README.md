# gdf - Go Disk Free

A enhanced `df` command written in Go to get free space for all mounted disks, displaying a textual gauge for each.

```
 ~/Projets/Go/jplozf/gdf>./gdf
/                             1.0 TB [############------------------] 41.89%
/home                         1.0 TB [############------------------] 41.89%
/boot                         1.0 GB [###########-------------------] 38.94%
/media/HDD                    2.0 TB [###---------------------------] 13.31%
/boot/efi                   627.9 MB [------------------------------]  3.22%
/media/WD001                 18.0 TB [#############-----------------] 45.93%
RAM                          33.3 GB [#########---------------------] 33.33%
```

By default, the gauges are displayed in color, as is the available RAM. These features can be modified using the following flags, which can be combined:
* `-m` to display the gauges in monochrome.
* `-d` to display only the disks.
