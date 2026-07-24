import { useState, useRef, useEffect } from "react";

export default function ModalRetirar({ item, aoFechar, aoConfirmar }) {
  const [quantidade, setQuantidade] = useState("");
  const [erro, setErro] = useState("");
  const [enviando, setEnviando] = useState(false);
  const campoRef = useRef(null);

  useEffect(() => {
    campoRef.current?.focus();
  }, []);

  async function aoEnviar(evento) {
    evento.preventDefault();
    const valor = Number(quantidade);

    if (!valor || valor <= 0) {
      setErro("Informe uma quantidade maior que zero.");
      return;
    }

    setEnviando(true);
    try {
      await aoConfirmar(valor);
      aoFechar();
    } catch (e) {
      setErro("Não foi possível registrar a retirada.");
    } finally {
      setEnviando(false);
    }
  }

  return (
    <div
      className="modal-fundo"
      role="dialog"
      aria-modal="true"
      aria-labelledby="titulo-modal-retirar"
      onClick={(e) => e.target === e.currentTarget && aoFechar()}
    >
      <div className="modal-caixa">
        <h2 id="titulo-modal-retirar">Retirar {item.nome}</h2>
        <p className="subtitulo">
          Em estoque: {item.quantidade} {item.unidade}
        </p>

        {erro && (
          <p className="mensagem-erro" role="alert">
            {erro}
          </p>
        )}

        <form onSubmit={aoEnviar} noValidate>
          <div className="campo-formulario">
            <label htmlFor="quantidade-retirar">
              Quantidade consumida ({item.unidade})
            </label>
            <input
              id="quantidade-retirar"
              ref={campoRef}
              type="number"
              min="0"
              step="0.01"
              required
              value={quantidade}
              onChange={(e) => setQuantidade(e.target.value)}
            />
          </div>

          <div className="acoes-modal">
            <button type="button" className="botao botao-secundario" onClick={aoFechar}>
              Cancelar
            </button>
            <button type="submit" className="botao botao-primario" disabled={enviando}>
              {enviando ? "Registrando..." : "Confirmar retirada"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
