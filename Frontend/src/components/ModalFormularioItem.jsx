import { useState, useRef, useEffect } from "react";
import { LOCAIS } from "./BarraFiltros.jsx";

// Mesmo modal serve para "adicionar" e "editar": se `item` vier preenchido,
// o formulário nasce com os valores dele; se vier null, nasce em branco.
export default function ModalFormularioItem({ item, aoFechar, aoSalvar }) {
  const ehEdicao = Boolean(item);

  const [nome, setNome] = useState(item?.nome ?? "");
  const [quantidade, setQuantidade] = useState(item?.quantidade ?? "");
  const [unidade, setUnidade] = useState(item?.unidade ?? "");
  const [local, setLocal] = useState(item?.local ?? LOCAIS[1]);
  const [estoqueMinimo, setEstoqueMinimo] = useState(item?.estoque_minimo ?? "");
  const [erro, setErro] = useState("");
  const [salvando, setSalvando] = useState(false);

  const primeiroCampoRef = useRef(null);

  // Foca automaticamente no primeiro campo ao abrir — ajuda quem navega
  // por teclado a não precisar dar Tab manualmente até o formulário.
  useEffect(() => {
    primeiroCampoRef.current?.focus();
  }, []);

  async function aoEnviar(evento) {
    evento.preventDefault();
    setErro("");

    if (!nome.trim() || !unidade.trim()) {
      setErro("Nome e unidade são obrigatórios.");
      return;
    }

    setSalvando(true);
    try {
      await aoSalvar({
        nome: nome.trim(),
        quantidade: Number(quantidade) || 0,
        unidade: unidade.trim(),
        local,
        estoque_minimo: Number(estoqueMinimo) || 0,
      });
      aoFechar();
    } catch (e) {
      setErro("Não foi possível salvar o item. Tente novamente.");
    } finally {
      setSalvando(false);
    }
  }

  return (
    // role="dialog" + aria-modal avisam leitores de tela que o resto da página
    // está temporariamente inacessível até esse modal fechar.
    <div
      className="modal-fundo"
      role="dialog"
      aria-modal="true"
      aria-labelledby="titulo-modal-item"
      onClick={(e) => e.target === e.currentTarget && aoFechar()}
    >
      <div className="modal-caixa">
        <h2 id="titulo-modal-item">{ehEdicao ? "Editar item" : "Adicionar item"}</h2>

        {erro && (
          <p className="mensagem-erro" role="alert">
            {erro}
          </p>
        )}

        <form onSubmit={aoEnviar} noValidate>
          <div className="campo-formulario">
            <label htmlFor="item-nome">Nome do item</label>
            <input
              id="item-nome"
              ref={primeiroCampoRef}
              type="text"
              required
              value={nome}
              onChange={(e) => setNome(e.target.value)}
            />
          </div>

          <div className="linha-dois-campos">
            <div className="campo-formulario">
              <label htmlFor="item-quantidade">Quantidade</label>
              <input
                id="item-quantidade"
                type="number"
                min="0"
                step="0.01"
                required
                value={quantidade}
                onChange={(e) => setQuantidade(e.target.value)}
              />
            </div>
            <div className="campo-formulario">
              <label htmlFor="item-unidade">Unidade</label>
              <input
                id="item-unidade"
                type="text"
                placeholder="kg, un, caixa..."
                required
                value={unidade}
                onChange={(e) => setUnidade(e.target.value)}
              />
            </div>
          </div>

          <div className="linha-dois-campos">
            <div className="campo-formulario">
              <label htmlFor="item-local">Localização</label>
              <select
                id="item-local"
                value={local}
                onChange={(e) => setLocal(e.target.value)}
              >
                {LOCAIS.slice(1).map((opcao) => (
                  <option key={opcao} value={opcao}>
                    {opcao}
                  </option>
                ))}
              </select>
            </div>
            <div className="campo-formulario">
              <label htmlFor="item-minimo">Estoque mínimo</label>
              <input
                id="item-minimo"
                type="number"
                min="0"
                step="0.01"
                value={estoqueMinimo}
                onChange={(e) => setEstoqueMinimo(e.target.value)}
              />
            </div>
          </div>

          <div className="acoes-modal">
            <button
              type="button"
              className="botao botao-secundario"
              onClick={aoFechar}
            >
              Cancelar
            </button>
            <button type="submit" className="botao botao-primario" disabled={salvando}>
              {salvando ? "Salvando..." : "Salvar"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
