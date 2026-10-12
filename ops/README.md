# Backup do banco (Controle de Estoque)

Backup diário do SQLite (`data/estoque.db`), às 04:00 UTC, por um timer do systemd.

- `backup-estoque.sh`: cópia consistente pela API de backup do SQLite, confere a integridade, compacta e envia.
- Destino: Oracle Object Storage (bucket `cadencia-backups`, prefixo `estoque/`) por um remote `crypt` do rclone (`estoque-crypt:`). Conteúdo e nomes saem da VPS já criptografados.
- Retenção de 30 dias: regra de ciclo de vida do bucket para o prefixo `estoque/`. A VPS não tem permissão de apagar objetos, de propósito.
- Nada em claro vai para o disco: o arquivo temporário fica em `/dev/shm`.
- As senhas do remote `crypt` ficam em `~/.config/rclone/rclone.conf` da VPS **e** em um gerenciador de senhas. Sem elas, os backups não abrem.

## Instalação (uma vez, na VPS)

```
sudo cp ops/systemd/estoque-backup.service ops/systemd/estoque-backup.timer /etc/systemd/system/
```

```
sudo systemctl daemon-reload && sudo systemctl enable --now estoque-backup.timer
```

Rodar na hora e conferir:

```
sudo systemctl start estoque-backup.service && journalctl -u estoque-backup.service -n 20 --no-pager
```

```
rclone lsl estoque-crypt:
```

## Restaurar

Baixar o backup mais recente (ou o que quiser), descompactar e conferir:

```
rclone copyto estoque-crypt:<nome>.db.gz /tmp/estoque.db.gz && gunzip /tmp/estoque.db.gz
```

```
python3 -c "import sqlite3;print(sqlite3.connect('/tmp/estoque.db').execute('pragma integrity_check').fetchone())"
```

Para voltar o dado: parar o container do backend, substituir `data/estoque.db` pelo arquivo restaurado (mantendo o dono e as permissões do arquivo original) e subir o container de novo.
