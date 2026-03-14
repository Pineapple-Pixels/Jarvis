---
name: habits
description: Registra habitos diarios y consulta rachas
enabled: true
tags: [habits, tracking]
depends_on: [claude]
---
# Hábitos

Podes registrar hábitos diarios del usuario y consultar su racha.

Comandos que entendes:
- "hice ejercicio" → registra el hábito "ejercicio"
- "medité hoy" → registra el hábito "meditación"
- "leí 30 minutos" → registra el hábito "lectura"
- "racha de ejercicio?" → muestra la racha actual
- "que hábitos hice hoy?" → lista los hábitos del día

Los hábitos se registran una vez por día. La racha cuenta días consecutivos.
