#!/usr/bin/env bash
# Backup diário do banco SQLite do Controle de Estoque (systemd: ops/systemd/estoque-backup.timer).
#
# - Cópia consistente pela API de backup do SQLite (segura mesmo com o app escrevendo).
# - Confere a integridade da cópia antes de enviar.
# - Envia para o Oracle Object Storage por um remote "crypt" do rclone: conteúdo e
#   nomes saem da VPS já criptografados.
# - Retenção de 30 dias: regra de ciclo de vida do bucket (prefixo estoque/). A VPS não
#   tem permissão de apagar objetos, de propósito.
# - O arquivo temporário fica em /dev/shm (memória): nada em claro vai para o disco.
set -euo pipefail

DATA_DIR="${ESTOQUE_DATA_DIR:-/home/ubuntu/apps/App-controle-de-estoque/data}"
REMOTE="${BACKUP_RCLONE_REMOTE:-estoque-crypt:}"
NAME="estoque_$(date -u +%Y%m%dT%H%M%SZ).db.gz"

TMP="$(mktemp -d /dev/shm/estoque-backup.XXXXXX)"
trap 'rm -rf "$TMP"' EXIT

# A pasta de dados pertence ao usuário do container, por isso o sudo. O arquivo de
# origem é aberto somente para leitura.
sudo -n python3 - "$DATA_DIR/estoque.db" "$TMP/estoque.db" <<'PY'
import sqlite3
import sys

origem = sqlite3.connect(f"file:{sys.argv[1]}?mode=ro", uri=True)
destino = sqlite3.connect(sys.argv[2])
origem.backup(destino)
resultado = destino.execute("PRAGMA integrity_check").fetchone()[0]
destino.close()
origem.close()
if resultado != "ok":
    sys.exit(f"integrity_check falhou: {resultado}")
PY

sudo -n chown "$(id -u):$(id -g)" "$TMP/estoque.db"
gzip -9 "$TMP/estoque.db"
rclone copyto "$TMP/estoque.db.gz" "${REMOTE}${NAME}"

echo "Backup enviado: ${NAME}"
