import { useEffect, useRef } from "react";
import { Html5Qrcode } from "html5-qrcode";

const ID_ELEMENTO_LEITOR = "leitor-qrcode-nota";

// Componente isolado de leitura de QR Code pela câmera. Fica responsável só
// por abrir a câmera, ler o código e devolver o texto lido — quem decide o
// que fazer com esse texto (validar domínio, consultar a SEFAZ, etc.) é o
// ModalImportarNota, que usa esse componente.
export default function LeitorQRCode({ aoLer, aoErro }) {
  const leitorRef = useRef(null);
  const emExecucaoRef = useRef(false);

  useEffect(() => {
    const leitor = new Html5Qrcode(ID_ELEMENTO_LEITOR);
    leitorRef.current = leitor;

    // Guardamos a PROMISE do start (não só disparamos e esquecemos). Isso é
    // importante por causa do StrictMode: em desenvolvimento, o React monta
    // o efeito, roda a limpeza e monta de novo — de propósito, pra pegar bug
    // de limpeza malfeita. Se a limpeza chamar stop() antes do start()
    // terminar de verdade, a lib lança "Cannot stop, scanner is not running"
    // de forma síncrona (não como rejeição de promise), então nem o .catch()
    // pega — some como erro não tratado no console e a câmera fica num
    // estado quebrado (tela preta). Encadeando o stop() DEPOIS do start()
    // resolver, isso nunca mais acontece — em produção (sem StrictMode)
    // esse encadeamento não muda nada visualmente, é transparente.
    const promessaInicio = leitor.start(
      { facingMode: "environment" }, // câmera traseira no celular
      { fps: 10, qrbox: { width: 240, height: 240 } },
      (textoLido) => {
        // Evita chamar aoLer mais de uma vez pro mesmo start (start
        // continua escaneando até ser parado explicitamente).
        if (emExecucaoRef.current) return;
        emExecucaoRef.current = true;
        pararCamera();
        aoLer(textoLido);
      },
      () => {
        // Callback de "não achou QR Code neste frame" — dispara a cada
        // frame sem leitura, então é esperado e não é um erro real.
      }
    );

    promessaInicio.catch(() => {
      aoErro(
        "Não foi possível acessar a câmera. Verifique a permissão do navegador ou cole o link manualmente abaixo."
      );
    });

    function pararCamera() {
      // Só chama stop() se o start() correspondente realmente teve sucesso
      // — se ele ainda não resolveu ou falhou, o .then() nem executa, e o
      // .catch() no final evita qualquer erro não tratado no console.
      promessaInicio
        .then(() => leitor.stop())
        .then(() => leitor.clear())
        .catch(() => {
          /* nunca chegou a iniciar, ou já foi parada — sem problema */
        });
    }

    return () => {
      // Ao desmontar (fechou o modal, trocou de passo, leu com sucesso),
      // para a câmera — senão ela continua ligada em segundo plano.
      pararCamera();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div
      id={ID_ELEMENTO_LEITOR}
      style={{
        width: "100%",
        maxWidth: "320px",
        margin: "0 auto",
        borderRadius: "8px",
        overflow: "hidden",
      }}
    />
  );
}
