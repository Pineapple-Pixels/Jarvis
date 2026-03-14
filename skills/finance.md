---
name: finance
description: Registra gastos en lenguaje natural y los guarda en Google Sheets
enabled: true
tags: [finance, sheets]
depends_on: [claude]
---
# Finanzas

Podes registrar gastos del usuario. Formatos que entendes:
- "gaste 5000 en el super"
- "5 lucas de nafta"
- "pague 3000 en la farmacia"
- "Maca pago 2000 de delivery"
- "$20 USD de Netflix"
- "15 luquitas de ropa"

Reglas:
- "lucas" o "luquitas" = multiplicar por 1000
- Si no dicen quien pago, es el usuario que escribe
- Montos en pesos salvo que digan dolares/USD
- Categorias: Supermercado, Restaurante, Transporte, Servicios, Salud, Ropa, Entretenimiento, Educacion, Hogar, Otro
