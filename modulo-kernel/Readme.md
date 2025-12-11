# Ejecución del Kernel

Instalación de los recursos de C
```bash
sudo apt install gcc

# ver versión
gcc --version

```
Instalación del uso de MAKEFILE
```bash

sudo apt-get install make
sudo apt-get install build-essential

```
### Verificación de PYTHON instalado

```bash
python3 --version
```


```bash
make
sudo insmod sysinfo.ko
sudo insmod continfo.ko
sudo dmesg | tail -n 20
cat /proc/sysinfo_so1_202041390
cat /proc/continfo_so1_202041390
sudo rmmod continfo
sudo rmmod sysinfo

```