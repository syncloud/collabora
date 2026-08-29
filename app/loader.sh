find_loader() {
  ROOT=$1

  case $(uname -m) in
    x86_64)
      TRIPLET=x86_64-linux-gnu
      LOADER=${ROOT}/usr/lib/${TRIPLET}/ld-linux-x86-64.so.2
      ;;
    aarch64)
      TRIPLET=aarch64-linux-gnu
      LOADER=${ROOT}/usr/lib/${TRIPLET}/ld-linux-aarch64.so.1
      ;;
    *)
      echo "unsupported architecture $(uname -m)" >&2
      exit 1
      ;;
  esac

  LIBS=${ROOT}/usr/lib/${TRIPLET}:${ROOT}/usr/lib:${ROOT}/lib:${ROOT}/lib64

  if [ ! -f "${LOADER}" ]; then
    echo "no dynamic loader at ${LOADER}" >&2
    exit 1
  fi
}
