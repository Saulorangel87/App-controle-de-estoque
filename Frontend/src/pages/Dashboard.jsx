import { useState, useEffect, useMemo, useCallback } from "react";
import { useAuth } from "../context/AuthContext.jsx";
import {
  listarItens,
  adicionarItem,
  editarItem,
  excluirItem,
  retirarItem,
} from "../api/api.js";
import CartoesResumo from "../components/CartoesResumo.jsx";
import BarraFiltros from "../components/BarraFiltros.jsx";
import TabelaItens from "../components/TabelaItens.jsx";
import ModalFormularioItem from "../components/ModalFormularioItem.jsx";
import ModalRetirar from "../components/ModalRetirar.jsx";

export default function Dashboard() {
  const { token, sair } = useAuth();

  const [itens, setItens] = useState([]);
  const [carregando, setCarregando] = useState(true);
  const [erro, setErro] = useState("");

  const [busca, setBusca] = useState("");
  const [filtroLocal, setFiltroLocal] = useState("");

  // Controla qual modal está aberto e com qual item (null = modal fechado / novo item)
  const [itemEmEdicao, setItemEmEdicao] = useState(undefined); // undefined = fechado
  const [itemParaRetirar, setItemParaRetirar] = useState(null);

  const carregarItens = useCallback(async () => {
    setCarregando(true);
    setErro("");
    try {
      const dados = await listarItens(token);
      setItens(dados);
    } catch (e) {
      setErro("Não foi possível carregar os itens do estoque.");
    } finally {
      setCarregando(false);
    }
  }, [token]);

  useEffect(() => {
    carregarItens();
  }, [carregarItens]);

  // useMemo evita recalcular a lista filtrada a cada renderização,
  // só refaz quando os itens, a busca ou o filtro realmente mudam.
  const itensFiltrados = useMemo(() => {
    return itens.filter((item) => {
      const bateBusca = item.nome
        .toLowerCase()
        .includes(busca.trim().toLowerCase());
      const bateLocal = filtroLocal ? item.local === filtroLocal : true;
      return bateBusca && bateLocal;
    });
  }, [itens, busca, filtroLocal]);

  const resumo = useMemo(() => {
    const emEstoqueBaixo = itens.filter(
      (item) => item.quantidade <= item.estoque_minimo
    ).length;
    const locaisUnicos = new Set(itens.map((item) => item.local));
    return {
      total: itens.length,
      estoqueBaixo: emEstoqueBaixo,
      locais: locaisUnicos.size,
    };
  }, [itens]);

  async function salvarItem(dadosItem) {
    if (itemEmEdicao) {
      await editarItem(itemEmEdicao.id, dadosItem, token);
    } else {
      await adicionarItem(dadosItem, token);
    }
    await carregarItens();
  }

  async function confirmarExclusao(item) {
    // window.confirm é suficiente aqui: ação destrutiva e pouco frequente,
    // não justifica construir mais um modal customizado só para isso.
    const confirmado = window.confirm(`Excluir "${item.nome}" do estoque?`);
    if (!confirmado) return;

    try {
      await excluirItem(item.id, token);
      await carregarItens();
    } catch (e) {
      setErro("Não foi possível excluir o item.");
    }
  }

  async function confirmarRetirada(quantidade) {
    await retirarItem(itemParaRetirar.id, quantidade, token);
    await carregarItens();
  }

  return (
    <main className="pagina">
      <div className="cabecalho-pagina">
        <div>
          <h1>Estoque</h1>
          <p className="subtitulo">{resumo.total} itens cadastrados</p>
        </div>
        <div style={{ display: "flex", gap: "8px" }}>
          <button
            type="button"
            className="botao botao-primario"
            onClick={() => setItemEmEdicao(null)}
          >
            + Item
          </button>
          <button type="button" className="botao botao-secundario" onClick={sair}>
            Sair
          </button>
        </div>
      </div>

      <CartoesResumo
        total={resumo.total}
        estoqueBaixo={resumo.estoqueBaixo}
        locais={resumo.locais}
      />

      <BarraFiltros
        busca={busca}
        aoMudarBusca={setBusca}
        local={filtroLocal}
        aoMudarLocal={setFiltroLocal}
      />

      {erro && (
        <p className="mensagem-erro" role="alert">
          {erro}
        </p>
      )}

      {/* aria-live avisa leitores de tela quando a lista termina de carregar,
          sem precisar que a pessoa fique navegando manualmente até descobrir */}
      <div aria-live="polite">
        {carregando ? (
          <p className="estado-vazio">Carregando itens...</p>
        ) : (
          <TabelaItens
            itens={itensFiltrados}
            aoEditar={setItemEmEdicao}
            aoExcluir={confirmarExclusao}
            aoRetirar={setItemParaRetirar}
          />
        )}
      </div>

      {itemEmEdicao !== undefined && (
        <ModalFormularioItem
          item={itemEmEdicao}
          aoFechar={() => setItemEmEdicao(undefined)}
          aoSalvar={salvarItem}
        />
      )}

      {itemParaRetirar && (
        <ModalRetirar
          item={itemParaRetirar}
          aoFechar={() => setItemParaRetirar(null)}
          aoConfirmar={confirmarRetirada}
        />
      )}
    </main>
  );
}
