# Architecture & Implementation Decisions (DECISIONS.md)

This document records architectural, technical, and implementation decisions made during the development of LabShop.

## 2026-10-12: Framework & Architecture Setup
- **Router**: Using `github.com/go-chi/chi/v5` as planned. Chi is lightweight, idiomatic, fully stdlib `http.Handler` compatible, and matches Clean Architecture requirements.
- **Frontend Stack**: Adopted PATCH v2.1: Vite + React 19 + TypeScript (strict) + Tailwind CSS + shadcn/ui. Go backend will not render HTML templates; `internal/adapter/web` is eliminated.
- **Directory Path with Special Characters**: Workspace contains `&` (`c:\CMC\WAF\WebForWAF&CDN`). All PowerShell scripts and Docker commands wrap paths in quotes and avoid using directory path as module/image tags.
- **Token Storage**: JWT access token kept in-memory, backed up in `sessionStorage` (wrapped with try/catch). `localStorage` is not used for auth tokens.
- **Agent Orchestration**: Multi-agent model established with Manager Agent coordinating specialized sub-agents: `backend_dev`, `frontend_dev`, and `qa_troubleshoot`.
