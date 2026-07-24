// Cards simples de resumo no topo do dashboard.
// Recebem os números já calculados pelo componente pai (Dashboard),
// então esse componente fica "burro" — só exibe, não calcula nada.
export default function CartoesResumo({ total, estoqueBaixo, locais }) {
  return (
    <section className="cartoes-resumo" aria-label="Resumo do estoque">
      <div className="cartao">
        <p className="cartao-rotulo">Itens cadastrados</p>
        <p className="cartao-valor">{total}</p>
      </div>
      <div className="cartao">
        <p className="cartao-rotulo">Estoque baixo</p>
        <p className={`cartao-valor ${estoqueBaixo > 0 ? "perigo" : ""}`}>
          {estoqueBaixo}
        </p>
      </div>
      <div className="cartao">
        <p className="cartao-rotulo">Localizações em uso</p>
        <p className="cartao-valor">{locais}</p>
      </div>
    </section>
  );
}
