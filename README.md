# API de productos y proveedores

API pequeña en Go, Gin, GORM y MySQL. Crea las tablas automáticamente al arrancar.

## Iniciar

```bash
docker compose up --build
```

La API queda en `http://localhost:8080`. Comprobar: `curl http://localhost:8080/health`.
MySQL está expuesto en el puerto 3306 para consultas locales. Los datos se guardan en el volumen `mysql_data`.
Las credenciales de `compose.yaml` son solo para desarrollo local; cámbialas antes de publicar el servicio.

## Endpoints

| Método | Ruta | Descripción |
| --- | --- | --- |
| GET, POST | `/api/proveedores` | Listar y crear proveedores |
| GET, PUT, DELETE | `/api/proveedores/:id` | Leer, reemplazar y borrar un proveedor |
| GET, POST | `/api/productos` | Listar y crear productos |
| GET, PUT, DELETE | `/api/productos/:id` | Leer, reemplazar y borrar un producto |
| GET | `/api/productos?proveedor_id=1` | Filtrar productos por proveedor |
| GET | `/api/consultas/resumen-proveedores` | JOIN y agregación por proveedor |
| GET | `/api/consultas/stock-bajo?limite=5` | JOIN de productos con stock bajo |
| GET | `/health` | Estado de API y MySQL |

`PUT` recibe el objeto completo. El email y el SKU deben ser únicos. No se puede borrar un proveedor con productos asociados.

## Ejemplo

```bash
curl -X POST http://localhost:8080/api/proveedores \
  -H 'Content-Type: application/json' \
  -d '{"nombre":"Acme","email":"ventas@acme.com"}'

curl -X POST http://localhost:8080/api/productos \
  -H 'Content-Type: application/json' \
  -d '{"sku":"P-001","nombre":"Teclado","precio":120.50,"stock":4,"proveedor_id":1}'

curl http://localhost:8080/api/consultas/resumen-proveedores
curl 'http://localhost:8080/api/consultas/stock-bajo?limite=5'
```

También se pueden ejecutar consultas SQL directamente:

```bash
docker compose exec db mysql -u productos -pproductos productos_db \
  -e 'SELECT p.nombre, p.stock, pr.nombre AS proveedor FROM productos p JOIN proveedores pr ON pr.id = p.proveedor_id;'
```

Para ejecutar sin Docker, configura `DB_DSN` con un DSN de MySQL que incluya `parseTime=True` y ejecuta `go run .`.
