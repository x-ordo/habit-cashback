# Actions에서 mTLS PEM 복원 스니펫

GitHub Secrets는 멀티라인도 가능하지만, **파일로 복원해서 쓰는 게 가장 안정적**입니다.

```yaml
- name: Write Toss mTLS certs
  shell: bash
  run: |
    mkdir -p /tmp/toss-mtls
    printf '%s' "${{ secrets.TOSS_MTLS_CERT_PEM }}" > /tmp/toss-mtls/cert.pem
    printf '%s' "${{ secrets.TOSS_MTLS_KEY_PEM }}"  > /tmp/toss-mtls/key.pem
    if [[ -n "${{ secrets.TOSS_MTLS_CA_PEM }}" ]]; then
      printf '%s' "${{ secrets.TOSS_MTLS_CA_PEM }}" > /tmp/toss-mtls/ca.pem
    fi

    echo "TOSS_MTLS_CERT_PATH=/tmp/toss-mtls/cert.pem" >> $GITHUB_ENV
    echo "TOSS_MTLS_KEY_PATH=/tmp/toss-mtls/key.pem"   >> $GITHUB_ENV
    if [[ -f /tmp/toss-mtls/ca.pem ]]; then
      echo "TOSS_MTLS_CA_PATH=/tmp/toss-mtls/ca.pem"   >> $GITHUB_ENV
    fi
```

Go 서버는 `TOSS_MTLS_CERT_PATH` / `TOSS_MTLS_KEY_PATH` 경로 기반으로 mTLS 설정을 잡도록 하는 게 운영이 편합니다.
(파일 기반이 컨테이너/서버리스에서도 제일 덜 꼬임)
