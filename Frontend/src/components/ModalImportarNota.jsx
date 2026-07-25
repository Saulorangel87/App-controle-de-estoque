import { useState, useRef } from "react";
import { importarNotaFiscal, confirmarImportacaoNota } from "../api/api.js";
import { LOCAIS } from "./BarraFiltros.jsx";

// Os 3 passos desse fluxo: escolher o arquivo, conferir os itens
// interpretados (e ajustar o que a comparação automática não resolveu
// sozinha), e um resumo final depois de aplicar.
const PASSOS = {
  UPLOAD: "upload",
  REVISAO: "revisao",
  SUCESSO: "sucesso",
};

export default function ModalImportarNota({ token, itensEstoque, aoFechar, aoConcluir }) {
  const [passo, setPasso] = useState(PASSOS.UPLOAD);
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState("");
  const [itensRevisao, setItensRevisao] = useState([]);
  const [resumo, setResumo] = useState(null);
  const inputArquivoRef = useRef(null);

  async function aoEnviarArquivo(evento) {
    evento.preventDefault();
    const arquivo = inputArquivoRef.current?.files?.[0];
    if (!arquivo) {
      setErro("Selecione o arquivo XML da nota fiscal.");
      return;
    }

    setErro("");
    setCarregando(true);
    try {
      const previa = await importarNotaFiscal(arquivo, token);
      // Cada item da prévia ganha campos extras editáveis (local, estoque
      // mínimo, nome final) só usados quando o status for "novo" — o
      // usuário pode ajustar antes de confirmar.
      const itensComEdicao = previa.map((item) => ({
        ...item,
        nomeFinal: item.nome_nota,
        local: LOCAIS[1],
        estoqueMinimo: 0,
      }));
      setItensRevisao(itensComEdicao);
      setPasso(PASSOS.REVISAO);
    } catch (e) {
      setErro(
        e.status === 400
          ? e.message || "Não foi possível ler esse XML."
          : "Erro ao processar o arquivo. Tente novamente."
      );
    } finally {
      setCarregando(false);
    }
  }

  function atualizarItem(indice, alteracoes) {
    setItensRevisao((atual) =>
      atual.map((item, i) => (i === indice ? { ...item, ...alteracoes } : item))
    );
  }

  // Troca um item de "novo" para "encontrado" quando o usuário reconhece
  // manualmente que ele já existe no estoque com outro nome.
  function associarAItemExistente(indice, itemIdTexto) {
    if (itemIdTexto === "") {
      atualizarItem(indice, {
        status: "novo",
        item_id: null,
        nome_atual: "",
        quantidade_atual: 0,
      });
      return;
    }
    const itemId = Number(itemIdTexto);
    const encontrado = itensEstoque.find((i) => i.id === itemId);
    if (!encontrado) return;
    atualizarItem(indice, {
      status: "encontrado",
      item_id: itemId,
      nome_atual: encontrado.nome,
      quantidade_atual: encontrado.quantidade,
    });
  }

  async function aoConfirmarImportacao() {
    setErro("");
    setCarregando(true);
    try {
      const entradas = itensRevisao.map((item) => ({
        item_id: item.status === "encontrado" ? item.item_id : null,
        nome: item.status === "encontrado" ? item.nome_atual : item.nomeFinal,
        quantidade: item.quantidade_nota,
        unidade: item.unidade_nota,
        local: item.local,
        estoque_minimo: Number(item.estoqueMinimo) || 0,
      }));

      const resultado = await confirmarImportacaoNota(entradas, token);
      setResumo(resultado);
      setPasso(PASSOS.SUCESSO);
    } catch (e) {
      setErro("Não foi possível aplicar a importação. Tente novamente.");
    } finally {
      setCarregando(false);
    }
  }

  function finalizar() {
    aoConcluir(); // recarrega a lista de itens no Dashboard
    aoFechar();
  }

  return (
    <div
      className="modal-fundo"
      role="dialog"
      aria-modal="true"
      aria-labelledby="titulo-modal-importar-nota"
      onClick={(e) => e.target === e.currentTarget && passo !== PASSOS.SUCESSO && aoFechar()}
    >
      <div className="modal-caixa modal-caixa-larga">
        <h2 id="titulo-modal-importar-nota">Importar nota fiscal</h2>

        {erro && (
          <p className="mensagem-erro" role="alert">
            {erro}
          </p>
        )}

        {passo === PASSOS.UPLOAD && (
          <form onSubmit={aoEnviarArquivo}>
            <p className="subtitulo" style={{ marginBottom: "12px" }}>
              Envie o arquivo XML da nota fiscal (NF-e). Os itens são lidos
              automaticamente e comparados com o seu estoque atual antes de
              qualquer alteração ser salva.
            </p>

            <div className="campo-formulario">
              <label htmlFor="arquivo-nota">Arquivo XML</label>
              <input
                id="arquivo-nota"
                type="file"
                accept=".xml,text/xml"
                ref={inputArquivoRef}
                required
              />
            </div>

            <div className="acoes-modal">
              <button type="button" className="botao botao-secundario" onClick={aoFechar}>
                Cancelar
              </button>
              <button type="submit" className="botao botao-primario" disabled={carregando}>
                {carregando ? "Processando..." : "Processar"}
              </button>
            </div>
          </form>
        )}

        {passo === PASSOS.REVISAO && (
          <>
            <p className="subtitulo" style={{ marginBottom: "12px" }}>
              Confira os itens antes de confirmar. Nada foi salvo ainda.
            </p>

            <div className="tabela-container" style={{ maxHeight: "360px", overflowY: "auto" }}>
              <table>
                <thead>
                  <tr>
                    <th scope="col">Item na nota</th>
                    <th scope="col">Qtd.</th>
                    <th scope="col">Situação</th>
                    <th scope="col">Ação</th>
                  </tr>
                </thead>
                <tbody>
                  {itensRevisao.map((item, indice) => (
                    <tr key={indice}>
                      <td>{item.nome_nota}</td>
                      <td>
                        {item.quantidade_nota} {item.unidade_nota}
                      </td>
                      <td>
                        {item.status === "encontrado" ? (
                          <span
                            className="badge-conta"
                            style={{
                              borderRadius: "6px",
                              width: "auto",
                              padding: "3px 8px",
                              fontSize: "12px",
                            }}
                          >
                            Encontrado
                          </span>
                        ) : (
                          <span
                            style={{
                              color: "var(--cor-texto-secundario)",
                              fontSize: "12px",
                              fontWeight: 600,
                            }}
                          >
                            Novo item
                          </span>
                        )}
                      </td>
                      <td>
                        {item.status === "encontrado" ? (
                          <div>
                            <p style={{ margin: 0, fontSize: "13px" }}>
                              {item.nome_atual}: {item.quantidade_atual} →{" "}
                              {(item.quantidade_atual + item.quantidade_nota).toFixed(2)}{" "}
                              {item.unidade_nota}
                            </p>
                            <button
                              type="button"
                              className="alternar-modo"
                              style={{ fontSize: "12px", marginTop: "4px" }}
                              onClick={() => associarAItemExistente(indice, "")}
                            >
                              Não é o mesmo item
                            </button>
                          </div>
                        ) : (
                          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
                            <select
                              aria-label={`Associar ${item.nome_nota} a um item existente`}
                              value=""
                              onChange={(e) => associarAItemExistente(indice, e.target.value)}
                            >
                              <option value="">Cadastrar como novo item</option>
                              {itensEstoque.map((existente) => (
                                <option key={existente.id} value={existente.id}>
                                  Já tenho: {existente.nome}
                                </option>
                              ))}
                            </select>
                            <select
                              aria-label={`Localização de ${item.nome_nota}`}
                              value={item.local}
                              onChange={(e) => atualizarItem(indice, { local: e.target.value })}
                            >
                              {LOCAIS.slice(1).map((local) => (
                                <option key={local} value={local}>
                                  {local}
                                </option>
                              ))}
                            </select>
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="acoes-modal">
              <button type="button" className="botao botao-secundario" onClick={aoFechar}>
                Cancelar
              </button>
              <button
                type="button"
                className="botao botao-primario"
                disabled={carregando}
                onClick={aoConfirmarImportacao}
              >
                {carregando ? "Aplicando..." : "Confirmar entrada"}
              </button>
            </div>
          </>
        )}

        {passo === PASSOS.SUCESSO && resumo && (
          <>
            <p style={{ color: "var(--cor-primaria)", fontWeight: 600 }}>
              Entrada realizada com sucesso!
            </p>
            <p className="subtitulo">
              {resumo.atualizados} {resumo.atualizados === 1 ? "item atualizado" : "itens atualizados"},{" "}
              {resumo.criados} {resumo.criados === 1 ? "item novo cadastrado" : "itens novos cadastrados"}.
            </p>
            <div className="acoes-modal">
              <button type="button" className="botao botao-primario" onClick={finalizar}>
                Fechar
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
