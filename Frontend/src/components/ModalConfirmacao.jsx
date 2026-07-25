import { useRef, useEffect } from "react";

// Modal de confirmação genérico — substitui window.confirm() por algo mais
// visível e difícil de confirmar sem querer no celular (o confirm() nativo
// do navegador é pequeno e o botão de confirmar fica perto do de cancelar,
// fácil de tocar errado). O botão de confirmar aqui nasce SEM foco automático
// de propósito — assim um toque duplo acidental não confirma a ação.
export default function ModalConfirmacao({
  titulo,
  mensagem,
  textoConfirmar = "Confirmar",
  aoCancelar,
  aoConfirmar,
}) {
  const botaoCancelarRef = useRef(null);

  // Foca no botão CANCELAR (não no de confirmar) — se a pessoa só apertar
  // Enter no reflexo, a ação segura (cancelar) é a que acontece.
  useEffect(() => {
    botaoCancelarRef.current?.focus();
  }, []);

  return (
    <div
      className="modal-fundo"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="titulo-modal-confirmacao"
      aria-describedby="mensagem-modal-confirmacao"
      onClick={(e) => e.target === e.currentTarget && aoCancelar()}
    >
      <div className="modal-caixa">
        <h2 id="titulo-modal-confirmacao">{titulo}</h2>
        <p id="mensagem-modal-confirmacao" className="subtitulo">
          {mensagem}
        </p>

        <div className="acoes-modal">
          <button
            ref={botaoCancelarRef}
            type="button"
            className="botao botao-secundario"
            onClick={aoCancelar}
          >
            Cancelar
          </button>
          <button
            type="button"
            className="botao botao-primario"
            style={{ backgroundColor: "var(--cor-perigo)" }}
            onClick={aoConfirmar}
          >
            {textoConfirmar}
          </button>
        </div>
      </div>
    </div>
  );
}
