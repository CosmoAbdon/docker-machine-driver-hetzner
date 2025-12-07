#!/bin/bash
set -e

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configurações (customize aqui)
VERSION_FILE="${VERSION_FILE:-.version}"
TAG_PREFIX="${TAG_PREFIX:-v}" #
REPO_NAME="${REPO_NAME:-$(basename $(git rev-parse --show-toplevel))}"

# Ler versão
if [ ! -f "$VERSION_FILE" ]; then
  echo -e "${RED}❌ Arquivo $VERSION_FILE não encontrado${NC}"
  exit 1
fi

VERSION=$(cat "$VERSION_FILE")
TAG="${TAG_PREFIX}${VERSION}"

echo -e "${BLUE}📦 Versão detectada: ${GREEN}$VERSION${NC}"
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
