# 🚀 Comandos de Deploy - Servidor SSH

## Backend (Go)

### Build
```bash
go build -o llantera-server ./cmd/server
```

### Ejecución con nohup
```bash
nohup ./llantera-server > server.log 2>&1 &
```

### Ver logs
```bash
tail -f server.log
```

### Detener servidor
```bash
pkill -f llantera-server
# o
ps aux | grep llantera-server
kill <PID>
```

---

## Frontend (Angular)

### Build para producción
```bash
cd ../trezo
ng build
```

### El resultado queda en:
```
dist/llantera-fe/
```

### Servir archivos estáticos (opción 1 - nginx)
Configurar nginx para servir desde `/path/to/project/dist/llantera-fe/`

### Servir archivos estáticos (opción 2 - simple server)
```bash
cd dist/llantera-fe
python3 -m http.server 8080
# o
npx serve -s . -p 8080
```

---

## Flujo Completo en SSH con Screen

### 1. Crear sesión screen
```bash
screen -S llantera
```

### 2. Backend
```bash
cd /path/to/llantera-project/llantera-hex
go build -o llantera-server ./cmd/server
nohup ./llantera-server > server.log 2>&1 &
```

### 3. Frontend
```bash
cd ../trezo
ng build
```

### 4. Verificar servicios
```bash
# Backend
ps aux | grep llantera-server
tail -f server.log

# Frontend (nginx)
sudo systemctl status nginx
sudo nginx -t
```

### 5. Salir de screen
```bash
Ctrl+A, D
```

### Reanudar screen
```bash
screen -r llantera
```

---

## 🔧 Ports Típicos

- **Backend API**: `8080`
- **Frontend**: `80` (nginx) o `8080` (serve)
- **Base de datos**: `5432` (PostgreSQL)

---

## 📝 Notas

- Asegurar que `.env` esté configurado antes del build
- Verificar permisos en archivos de `uploads/`
- Para producción, considerar usar systemd en lugar de nohup
- Los logs del backend se guardan en `server.log`
