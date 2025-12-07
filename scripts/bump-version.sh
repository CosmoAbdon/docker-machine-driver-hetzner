#!/bin/bash
set -e

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

VERSION_FILE="${VERSION_FILE:-.version}"

increment_version() {
  local version=$1
  local type=$2

  version=${version#v}

  IFS='.' read -r -a parts <<<"$version"
  major="${parts[0]}"
  minor="${parts[1]}"
  patch="${parts[2]}"

  case $type in
  patch)
    patch=$((patch + 1))
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  esac

  echo "v$major.$minor.$patch"
}

if [ -f "$VERSION_FILE" ]; then
  CURRENT_VERSION=$(cat "$VERSION_FILE")
else
  echo -e "${YELLOW}⚠️  Arquivo $VERSION_FILE não encontrado${NC}"
  echo -e "${BLUE}💡 Criando com versão inicial v0.1.0${NC}"
  CURRENT_VERSION="v0.1.0"
  echo "$CURRENT_VERSION" >"$VERSION_FILE"
fi

if [[ ! $CURRENT_VERSION =~ ^v ]]; then
  CURRENT_VERSION="v$CURRENT_VERSION"
fi

echo -e "${BLUE}📦 Versão atual: ${GREEN}$CURRENT_VERSION${NC}"
echo ""
echo "Escolha o tipo de atualização:"
echo ""
echo -e "  ${YELLOW}1)${NC} patch  → $(increment_version $CURRENT_VERSION patch) ${GREEN}(correções de bugs)${NC}"
echo -e "  ${YELLOW}2)${NC} minor  → $(increment_version $CURRENT_VERSION minor) ${GREEN}(novas funcionalidades)${NC}"
echo -e "  ${YELLOW}3)${NC} major  → $(increment_version $CURRENT_VERSION major) ${GREEN}(breaking changes)${NC}"
echo -e "  ${YELLOW}4)${NC} custom → ${GREEN}(versão personalizada)${NC}"
echo -e "  ${RED}0)${NC} cancelar"
echo ""
read -p "Opção: " option

case $option in
1)
  NEW_VERSION=$(increment_version $CURRENT_VERSION patch)
  ;;
2)
  NEW_VERSION=$(increment_version $CURRENT_VERSION minor)
  ;;
3)
  NEW_VERSION=$(increment_version $CURRENT_VERSION major)
  ;;
4)
  read -p "Digite a nova versão (com ou sem 'v'): " NEW_VERSION
  if [[ ! $NEW_VERSION =~ ^v ]]; then
    NEW_VERSION="v$NEW_VERSION"
  fi
  ;;
0)
  echo -e "${RED}❌ Cancelado${NC}"
  exit 0
  ;;
*)
  echo -e "${RED}❌ Opção inválida${NC}"
  exit 1
  ;;
esac

if ! [[ $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo -e "${RED}❌ Formato de versão inválido. Use: vX.Y.Z (exemplo: v1.0.0)${NC}"
  exit 1
fi

echo ""
echo -e "${BLUE}📦 Nova versão: ${GREEN}$NEW_VERSION${NC}"
echo ""
read -p "Confirma a atualização? (y/n): " confirm

if [[ $confirm != "y" ]]; then
  echo -e "${RED}❌ Cancelado${NC}"
  exit 0
fi

echo ""
echo -e "${BLUE}📝 Atualizando $VERSION_FILE...${NC}"
echo "$NEW_VERSION" >"$VERSION_FILE"

echo -e "${BLUE}🔄 Commitando mudanças...${NC}"
git add "$VERSION_FILE"
git commit -m "chore: bump version to $NEW_VERSION"

echo ""
echo -e "${GREEN}✅ Versão atualizada para: $NEW_VERSION${NC}"
echo ""
echo -e "${YELLOW}Deseja criar release agora? (y/n):${NC} "
read -r release_confirm

if [[ $release_confirm == "y" ]]; then
  if [ -f "./release.sh" ]; then
    ./release.sh
  else
    echo -e "${YELLOW}⚠️  Script release.sh não encontrado${NC}"
    echo -e "${BLUE}💡 Execute manualmente: ./release.sh${NC}"
  fi
else
  echo ""
  echo -e "${GREEN}✅ Pronto!${NC}"
  echo -e "${BLUE}💡 Para criar o release depois, execute: ./release.sh${NC}"
fi
