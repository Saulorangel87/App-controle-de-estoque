import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { login, cadastrar } from "../api/api.js";
import { useAuth } from "../context/AuthContext.jsx";

export default function Login() {
  // Alterna entre "entrar" (login) e "criar conta" (cadastro) na mesma tela,
  // evitando duas rotas separadas para um formulário praticamente igual.
  const [modoCadastro, setModoCadastro] = useState(false);
  const [nome, setNome] = useState("");
  const [senha, setSenha] = useState("");
  const [erro, setErro] = useState("");
  const [carregando, setCarregando] = useState(false);

  const { entrar } = useAuth();
  const navegar = useNavigate();

  async function aoEnviar(evento) {
    evento.preventDefault();
    setErro("");
    setCarregando(true);

    try {
      if (modoCadastro) {
        // Depois de criar a conta, já faz login automaticamente para não
        // obrigar o usuário a digitar tudo de novo.
        await cadastrar(nome, senha);
        const resultado = await login(nome, senha);
        entrar(resultado.token);
      } else {
        const resultado = await login(nome, senha);
        entrar(resultado.token);
      }
      navegar("/");
    } catch (e) {
      setErro(
        modoCadastro
          ? "Não foi possível criar a conta. O nome de usuário já pode estar em uso."
          : "Usuário ou senha inválidos."
      );
    } finally {
      setCarregando(false);
    }
  }

  return (
    <main className="pagina-login">
      <div className="caixa-login">
        <h1>{modoCadastro ? "Criar conta" : "Entrar"}</h1>
        <p className="subtitulo">Controle de estoque de mantimentos de casa</p>

        {erro && (
          <p className="mensagem-erro" role="alert">
            {erro}
          </p>
        )}

        <form onSubmit={aoEnviar} noValidate>
          <div className="campo-formulario">
            <label htmlFor="campo-nome">Nome de usuário</label>
            <input
              id="campo-nome"
              name="nome"
              type="text"
              autoComplete="username"
              required
              value={nome}
              onChange={(e) => setNome(e.target.value)}
            />
          </div>

          <div className="campo-formulario">
            <label htmlFor="campo-senha">Senha</label>
            <input
              id="campo-senha"
              name="senha"
              type="password"
              autoComplete={modoCadastro ? "new-password" : "current-password"}
              required
              minLength={6}
              value={senha}
              onChange={(e) => setSenha(e.target.value)}
            />
          </div>

          <button
            type="submit"
            className="botao botao-primario"
            style={{ width: "100%", justifyContent: "center" }}
            disabled={carregando}
          >
            {carregando
              ? "Aguarde..."
              : modoCadastro
              ? "Criar conta"
              : "Entrar"}
          </button>
        </form>

        <button
          type="button"
          className="alternar-modo"
          onClick={() => {
            setModoCadastro((atual) => !atual);
            setErro("");
          }}
        >
          {modoCadastro
            ? "Já tenho conta — entrar"
            : "Ainda não tenho conta — Criar"}
        </button>
      </div>
    </main>
  );
}
