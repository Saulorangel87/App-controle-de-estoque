const LOCAIS = ["Todos os locais", "Despensa", "Geladeira", "Freezer", "Armário"];

export default function BarraFiltros({
  busca,
  aoMudarBusca,
  local,
  aoMudarLocal,
  apenasEstoqueBaixo,
  aoMudarApenasEstoqueBaixo,
}) {
  return (
    <div className="barra-filtros">
      <div className="campo-busca">
        {/* Label associado por htmlFor, visível só para leitor de tela (sr-only),
            já que o placeholder do input dá o contexto visual */}
        <label htmlFor="busca-item" className="sr-only">
          Buscar item pelo nome
        </label>
        <input
          id="busca-item"
          type="search"
          placeholder="Buscar item..."
          value={busca}
          onChange={(e) => aoMudarBusca(e.target.value)}
        />
      </div>

      <div>
        <label htmlFor="filtro-local" className="sr-only">
          Filtrar por localização
        </label>
        <select
          id="filtro-local"
          value={local}
          onChange={(e) => aoMudarLocal(e.target.value)}
        >
          {LOCAIS.map((opcao) => (
            <option key={opcao} value={opcao === "Todos os locais" ? "" : opcao}>
              {opcao}
            </option>
          ))}
        </select>
      </div>

      {/* Botão de alternância (não checkbox escondido) para ficar visualmente
          consistente com os outros botões da barra de filtros */}
      <button
        type="button"
        className={`botao ${apenasEstoqueBaixo ? "botao-primario" : "botao-secundario"}`}
        onClick={() => aoMudarApenasEstoqueBaixo(!apenasEstoqueBaixo)}
        aria-pressed={apenasEstoqueBaixo}
      >
        <span aria-hidden="true">⚠️</span> Estoque baixo
      </button>
    </div>
  );
}

export { LOCAIS };
