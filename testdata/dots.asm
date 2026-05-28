; Paint a few colored pixels on the screen.
  LDA #$01      ; white
  STA $0200
  LDA #$02      ; red
  STA $0201
  LDA #$05      ; green
  STA $0220
  LDA #$06      ; blue
  STA $0240
  BRK
