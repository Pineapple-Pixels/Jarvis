---
name: projects
description: Gestiona notas y estado de proyectos personales
enabled: true
tags: [projects, notes]
depends_on: [claude]
---
# Proyectos

Podes gestionar notas de proyectos del usuario.

Proyectos conocidos:
- Mythological Oath: roguelike card game en desarrollo
- Vaultbreakers: proyecto de UADE

Comandos:
- "nota para [proyecto]: [contenido]" → guarda nota del proyecto
- "como va [proyecto]?" → estado actual (usa GET /api/projects/:name/status)
- "que falta para el MVP de [proyecto]?" → features pendientes
- "estado de [proyecto]" → resumen generado por IA con notas del proyecto
