# Acceso al Dashboard

1. Abrir un navegador web
2. Ingresar a: http://localhost:3000
3. Usuario: admin
4. Contraseña: admin
5. Configuración de la base de datos.
**Paso 1** Desde el panel de grafana seleccionar en donde dice *Add your first data source*
![Panel de Grafana](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/panel_grafana.png)

**Paso 2** Seleccionar SQLITE
![seleccion base de datos](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/configuracion_bd.png)

**Paso 3** Configuración para la conexión a la Base De Datos.
![Configuracion de BD](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/config_db.png)
```bash
/var/lib/sqlite/monitor.db

mode=ro&_ignore_check_constraints=1
```
Luego de llenar los campos, dar *Save & test*

6. Seleccionar el dashboard del proyecto
**Paso 1** Desde el panel de grafana seleccionar en el menu de grafana, en donde dice *Home*
![Menu de Grafana](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/configuracion_dashboard.png)

Y seleccionamos en donde dice *Dashboards*

**Paso 2** Le damos click al boton de *New* y le damos a *Import*
![Dashboards](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/config_dashboard.png)

**Paso 3** Le damos click a donde dice *Upload dashboard JSON file*, aqui nos vamos a la carpeta de dashboard de nuestro proyecto, y seleccionamos el archivo *dashboard.json*
![Import](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/importar_json.png)
Luego se procede a dar click a *Load* y ya estaria conectado al Dashboard