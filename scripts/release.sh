#!/bin/bash
set -e

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

VERSION_FILE="${VERSION_FILE:-.version}"

if [ ! -f "$VERSION_FILE" ]; then
  echo -e "${RED}❌ Arquivo $VERSION_FILE não encontrado${NC}"
  echo -e "${BLUE}💡 Execute primeiro: ./bump-version-generic.sh${NC}"
  exit 1
fi

TAG=$(cat "$VERSION_FILE")

if ! [[ $TAG =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo -e "${RED}❌ Formato de versão inválido em $VERSION_FILE${NC}"
  echo -e "${BLUE}💡 Esperado: vX.Y.Z (exemplo: v1.0.0)${NC}"
  exit 1
fi

echo -e "${BLUE}📦 Versão: ${GREEN}$TAG${NC}"
echo -e "${BLUE}🏷️  Tag que será criada: ${GREEN}$TAG${NC}"
echo ""

echo -e "${YELLOW}🗑️  Deletando tag antiga (se existir)...${NC}"
git tag -d "$TAG" 2>/dev/null || true
git push origin ":refs/tags/$TAG" 2>/dev/null || true

echo -e "${YELLOW}🗑️  Deletando release antigo (se existir)...${NC}"
gh release delete "$TAG" --yes 2>/dev/null || true

echo -e "${BLUE}🏷️  Criando nova tag...${NC}"
git tag "$TAG"

echo -e "${BLUE}🚀 Pushing tag...${NC}"
git push origin "$TAG"

echo ""
echo -e "${GREEN}✅ Tag $TAG criada com sucesso!${NC}"
echo -e "${BLUE}💡 GitHub Actions criará o release automaticamente.${NC}"
