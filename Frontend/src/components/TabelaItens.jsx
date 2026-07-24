export default function TabelaItens({ itens, aoEditar, aoExcluir, aoRetirar }) {
  if (itens.length === 0) {
    return (
      <div className="tabela-container">
        <p className="estado-vazio">
          Nenhum item encontrado. Adicione o primeiro item do seu estoque.
        </p>
      </div>
    );
  }

  return (
    <div className="tabela-container">
      <table>
        <caption className="sr-only">Lista de itens do estoque</caption>
        <thead>
          <tr>
            <th scope="col">Item</th>
            <th scope="col">Local</th>
            <th scope="col">Quantidade</th>
            <th scope="col">
              <span className="sr-only">Ações</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {itens.map((item) => {
            const estoqueBaixo = item.quantidade <= item.estoque_minimo;
            return (
              <tr key={item.id}>
                <td>{item.nome}</td>
                <td>{item.local}</td>
                <td>
                  {estoqueBaixo ? (
                    <span className="badge-estoque-baixo">
                      {item.quantidade} {item.unidade}
                      {/* aria-hidden porque o texto já diz que está baixo — o emoji é só reforço visual */}
                      <span aria-hidden="true">⚠️</span>
                    </span>
                  ) : (
                    <span>
                      {item.quantidade} {item.unidade}
                    </span>
                  )}
                </td>
                <td>
                  <div className="acoes-linha">
                    <button
                      type="button"
                      className="botao botao-secundario"
                      onClick={() => aoRetirar(item)}
                    >
                      Retirar
                    </button>
                    <button
                      type="button"
                      className="botao botao-icone"
                      onClick={() => aoEditar(item)}
                      aria-label={`Editar ${item.nome}`}
                    >
                      ✏️
                    </button>
                    <button
                      type="button"
                      className="botao botao-icone botao-perigo"
                      onClick={() => aoExcluir(item)}
                      aria-label={`Excluir ${item.nome}`}
                    >
                      🗑️
                    </button>
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
