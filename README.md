# Auth GO
Este é um projeto simples de autenticador em Go que permite o registro e a autenticação de usuários, seguindo um fluxo simples de registro, login e logout.

## Funcionalidades
* Registro de Usuário: Permite que novos usuários se cadastrem no sistema informando um nome de usuário e uma senha.
As senhas são armazenadas de forma segura utilizando técnicas de hash para garantir a segurança dos dados.
* Autenticação (Login): Usuários registrados podem fazer login fornecendo seu nome de usuário e senha.
Se as credenciais estiverem corretas, um token de sessão é gerado para autenticação futura.
* Logout: O sistema permite que os usuários encerrem sua sessão removendo a sessão ativa.
* Persistência de Dados: Os usuários são armazenados em um banco de dados PostgreSQL e as sessões ativas são armazenadas em Redis. 
* Segurança: Senhas são armazenadas usando hash seguro (bcrypt).
O sistema pode implementar proteção contra ataques como força bruta e SQL Injection.

## Requisitos

* Go
* PostgreSQL
* Redis
* Docker

