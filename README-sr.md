# last.fm plejer
[README in english](README.md). <br>
Projekat za predmet Programske paradigme na Matematickom fakultetu Univerziteta u Beogradu. <br>
GUI muzicki plejer napisan u programskom jeziku Go koji koristi [last.fm](https://www.last.fm/) API da skine korisnikov `mix.json`, parsira ga i skida i pusta pesme. <br>
**Jos uvek u izradi**.

## Preostali posao
U pribliznom redosledu implementiranja:
- [x] download thread
- [x] player thread
- [x] poliranje download i player thread funkcija
- [x] poliranje double list strukture i funkcija
- [x] skidanje unapred
- [x] skidanje i parsiranje `mix.json` fajla za konkretnog korisnika
- [x] automatsko pustanje naredne pesme nakon sto se trenutna zavrsi
- [x] seeking
- [x] kontrole na tastaturi
- [x] informacije o pesmi koja se trenutno pusta: ime pesme, izvodjaca, vreme trajanja/ukupno vreme, slika albuma
- [x] pomeranje liste u levo i postavljanje informacija o pesmi desno, sa dugmicima za kontrolu ispod
- [x] kontrole za jacinu zvuka
- [ ] prikaz zvuka
- [ ] panel sa podesavanjima
- [ ] prikazi progres skidanja i indikator za skinute pesme
- [ ] opcija za skrivanje i prikazivanje plejliste
- [ ] osvezavanje plejliste (uz brisanje starih pesama)
- [ ] skidanje novog miksa unapred i nalepljivanje na kraj trenutnog
<br>

*Mozda*

- [ ] prerada kontrole na tastaturi (za koriscenje bez mod tastera)
- [ ] vizuelizacija
- [ ] automatsko brisanje pesama
- [ ] direktno pustanje opus/webm fajlova (ukoliko je moguce)
- [ ] skroblovanje na last.fm