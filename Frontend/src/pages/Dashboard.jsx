import { useState, useEffect, useMemo, useCallback } from "react";
import { useAuth } from "../context/AuthContext.jsx";
import { useTheme } from "../context/ThemeContext.jsx";
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
  const { token, nome, sair } = useAuth();
  const { tema, alternarTema } = useTheme();

  const [itens, setItens] = useState([]);
  const [carregando, setCarregando] = useState(true);
  const [erro, setErro] = useState("");

  const [busca, setBusca] = useState("");
  const [filtroLocal, setFiltroLocal] = useState("");
  const [apenasEstoqueBaixo, setApenasEstoqueBaixo] = useState(false);

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

  // Mostra só o primeiro e o segundo nome da conta no cabeçalho (ex: "Saulo Rangel
  // Silva" vira "Saulo Rangel"), para não estourar o layout com nomes longos.
  const nomeExibicao = useMemo(() => {
    const partes = nome.trim().split(/\s+/).filter(Boolean);
    return partes.slice(0, 2).join(" ");
  }, [nome]);

  const iniciaisConta = useMemo(() => {
    const partes = nome.trim().split(/\s+/).filter(Boolean);
    return partes
      .slice(0, 2)
      .map((parte) => parte[0])
      .join("")
      .toUpperCase();
  }, [nome]);

  // useMemo evita recalcular a lista filtrada a cada renderização,
  // só refaz quando os itens ou algum dos filtros realmente mudam.
  const itensFiltrados = useMemo(() => {
    return itens.filter((item) => {
      const bateBusca = item.nome
        .toLowerCase()
        .includes(busca.trim().toLowerCase());
      const bateLocal = filtroLocal ? item.local === filtroLocal : true;
      const bateEstoqueBaixo = apenasEstoqueBaixo
        ? item.quantidade <= item.estoque_minimo
        : true;
      return bateBusca && bateLocal && bateEstoqueBaixo;
    });
  }, [itens, busca, filtroLocal, apenasEstoqueBaixo]);

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

        <div className="acoes-cabecalho">
          <button
            type="button"
            className="botao botao-primario"
            onClick={() => setItemEmEdicao(null)}
          >
            + Item
          </button>

          <button
            type="button"
            className="botao-tema"
            onClick={alternarTema}
            aria-label={
              tema === "claro" ? "Mudar para tema escuro" : "Mudar para tema claro"
            }
          >
            <span aria-hidden="true">{tema === "claro" ? "🌙" : "☀️"}</span>
          </button>

          {nomeExibicao && (
            <span
              className="badge-conta"
              title={nomeExibicao}
              aria-label={`Conta: ${nomeExibicao}`}
            >
              {iniciaisConta}
            </span>
          )}

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
        apenasEstoqueBaixo={apenasEstoqueBaixo}
        aoMudarApenasEstoqueBaixo={setApenasEstoqueBaixo}
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
