# gdf - Go Disk Free

A enhanced `df` command written in Go to get free space for all mounted disks, RAM and CPU usage, displaying a textual gauge for each.
```
 ~/Projets/Go/jplozf/gdf>./gdf
/                             1.0 TB [###########-------------------] 39.43%
/boot                         1.0 GB [###########-------------------] 38.99%
/home                         1.0 TB [###########-------------------] 39.43%
/media/HDD                    2.0 TB [####--------------------------] 15.03%
/boot/efi                   627.9 MB [------------------------------]  3.22%
/media/WD001                 18.0 TB [#############-----------------] 46.19%
RAM                          33.3 GB [########----------------------] 29.62%
CPU                             1 mn [#######-----------------------] 23.56%
CPU                             5 mn [#########---------------------] 31.69%
CPU                            15 mn [########----------------------] 28.44%
```
By default, the gauges are displayed in color, as is the available RAM and CPU. These features can be modified using the following flags, which can be combined:
```
Usage of gdf:
  -a    Display all metrics
  -c    Display CPU metrics
  -d    Display file systems metrics
  -m    Display output in monochrome without colors
  -r    Display RAM metrics
  -w int
        Watch every n seconds
```
The `-w` flag followed by an integer value `n` starts a continuous display where values are refreshed every `n` seconds. Just press `Ctrl+C` to stop the loop.