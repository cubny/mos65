;  _____         _
; |   __|___ ___| |_ ___
; |__   |   | .'| '_| -_|
; |_____|_|_|__,|_,_|___|
;
; Change direction with W A S D.

define appleL         $00
define appleH         $01
define snakeHeadL     $10
define snakeHeadH     $11
define snakeBodyStart $12
define snakeDirection $02
define snakeLength    $03

define movingUp      1
define movingRight   2
define movingDown    4
define movingLeft    8

define ASCII_w 119
define ASCII_a 97
define ASCII_s 115
define ASCII_d 100

define sysRandom $fe
define sysLastKey $ff

  jsr init
  jsr loop

init:
  jsr initSnake
  jsr generateApplePosition
  rts

initSnake:
  lda #movingRight
  sta snakeDirection
  lda #4
  sta snakeLength
  lda #$11
  sta snakeHeadL
  lda #$10
  sta snakeBodyStart
  lda #$0f
  sta $14
  lda #$04
  sta snakeHeadH
  sta $13
  sta $15
  rts

generateApplePosition:
  lda sysRandom
  sta appleL
  lda sysRandom
  and #$03
  clc
  adc #2
  sta appleH
  rts

loop:
  jsr readKeys
  jsr checkCollision
  jsr updateSnake
  jsr drawApple
  jsr drawSnake
  jsr spinWheels
  jmp loop

readKeys:
  lda sysLastKey
  cmp #ASCII_w
  beq upKey
  cmp #ASCII_d
  beq rightKey
  cmp #ASCII_s
  beq downKey
  cmp #ASCII_a
  beq leftKey
  rts
upKey:
  lda #movingDown
  bit snakeDirection
  bne illegalMove
  lda #movingUp
  sta snakeDirection
  rts
rightKey:
  lda #movingLeft
  bit snakeDirection
  bne illegalMove
  lda #movingRight
  sta snakeDirection
  rts
downKey:
  lda #movingUp
  bit snakeDirection
  bne illegalMove
  lda #movingDown
  sta snakeDirection
  rts
leftKey:
  lda #movingRight
  bit snakeDirection
  bne illegalMove
  lda #movingLeft
  sta snakeDirection
  rts
illegalMove:
  rts

checkCollision:
  jsr checkAppleCollision
  jsr checkSnakeCollision
  rts

checkAppleCollision:
  lda appleL
  cmp snakeHeadL
  bne doneCheckingAppleCollision
  lda appleH
  cmp snakeHeadH
  bne doneCheckingAppleCollision
  inc snakeLength
  inc snakeLength
  jsr generateApplePosition
doneCheckingAppleCollision:
  rts

checkSnakeCollision:
  ldx #2
snakeCollisionLoop:
  lda snakeHeadL,x
  cmp snakeHeadL
  bne continueCollisionLoop
maybeCollided:
  lda snakeHeadH,x
  cmp snakeHeadH
  beq didCollide
continueCollisionLoop:
  inx
  inx
  cpx snakeLength
  beq didntCollide
  jmp snakeCollisionLoop
didCollide:
  jmp gameOver
didntCollide:
  rts

updateSnake:
  ldx snakeLength
  dex
  txa
updateloop:
  lda snakeHeadL,x
  sta snakeBodyStart,x
  dex
  bpl updateloop
  lda snakeDirection
  lsr a
  bcs up
  lsr a
  bcs right
  lsr a
  bcs down
  lsr a
  bcs left
up:
  lda snakeHeadL
  sec
  sbc #$20
  sta snakeHeadL
  bcc upup
  rts
upup:
  dec snakeHeadH
  lda #$1
  cmp snakeHeadH
  beq collision
  rts
right:
  inc snakeHeadL
  lda #$1f
  bit snakeHeadL
  beq collision
  rts
down:
  lda snakeHeadL
  clc
  adc #$20
  sta snakeHeadL
  bcs downdown
  rts
downdown:
  inc snakeHeadH
  lda #$6
  cmp snakeHeadH
  beq collision
  rts
left:
  dec snakeHeadL
  lda snakeHeadL
  and #$1f
  cmp #$1f
  beq collision
  rts
collision:
  jmp gameOver

drawApple:
  ldy #0
  lda sysRandom
  sta (appleL),y
  rts

drawSnake:
  ldx snakeLength
  lda #0
  sta (snakeHeadL,x)
  ldx #0
  lda #1
  sta (snakeHeadL,x)
  rts

spinWheels:
  ldx #0
spinloop:
  nop
  nop
  dex
  bne spinloop
  rts

gameOver: