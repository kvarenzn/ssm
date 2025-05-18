for f in *.txt; do
  mv -- "$f" "${f%%_*}_${f##*_}"
done
