YAML файлы добавлены для наглядности и читаемости.
Данные грузятся из json'ов.

Проверить результат можно командой:
```bash
yc compute instance create \
  --folder-id $folderId \
  --name test \
  --zone ru-central1-c \
  --cores 2 \
  --memory 2gb \
  --network-interface subnet-name=default-ru-central1-c,nat-ip-version=ipv4\
  --create-boot-disk image-id=fd8h059qvtg37ks9ke9o \
  --metadata-from-file user-data=result/user-data \
  --profile default
```