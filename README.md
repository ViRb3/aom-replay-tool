## Age of Mythology Retold Tool

### Features

- Unpack a replay from its "l33t" compressed data, so it can be explored and hex edited in-place
- Repack a previously unpacked replay into a fully functional replay that can be read by the game
- 010 Editor template to parse an unpacked file, including:
  - Nodes
    - XMBs
    - Profile keys
    - Build info
    - Heuristics for unknown node traversal
  - Game commands

### Resources

- https://github.com/erin-fitzpatric/next-aom-gg/tree/main/src/server/recParser
- https://github.com/jerkeeler/restoration
- https://github.com/Logg-y/retoldrecprocessor

### Future plans

Anonymize replay from profile keys:

- gameplayfabpartyaddress
- gameplayer1rlinkid
- gameplayer1name

- gameplayer1pfentity