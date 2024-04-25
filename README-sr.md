# last.fm plejer
[README in english](README.md). <br>
Projekat za predmet Programske paradigme na Matematickom fakultetu Univerziteta u Beogradu. <br>
GUI muzicki plejer napisan u programskom jeziku Go koji koristi [last.fm](https://www.last.fm/) API da skine korisnikov `mix.json`, parsira ga i skida i pusta pesme. <br>
**Jos uvek u izradi**.

## Precice
- mod taster: **control**
- **q** ugasi program
- **space** pusti/pauziraj
- **;** prethodna pesma
- **'** naredna pesma
- **-** smanji zvuk
- **=** pojacaj zvuk
- **m** mutiraj
- **,** skoci 5 sekundi unazad
- **.** skoci 5 sekundi unapred
- **p** sakrij/prikazi listu pesama


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
- [x] prikaz zvuka (slajder za zvuk)
- [x] panel sa podesavanjima
- [x] pri logovanju/ulasku u program, odmah skini prvu pesmu
- [ ] prikazi progres skidanja i indikator za skinute pesme
- [x] opcija za skrivanje i prikazivanje plejliste
- [ ] osvezavanje plejliste (uz brisanje starih pesama)
- [ ] skidanje novog miksa unapred i nalepljivanje na kraj trenutnog
<br>

*Mozda*

- [ ] prerada kontrole na tastaturi (za koriscenje bez mod tastera)
- [ ] skroblovanje na last.fm