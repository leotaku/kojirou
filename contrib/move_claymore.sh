#!/usr/bin/env sh

cd Claymore

move() {
    target="$1"
    mkdir -p "$target"
    shift
    for number in "$@"; do
        mv Unknown/"$(printf "%04d" "$number")" "$target"
    done
}

move 01  1  2  3  4
move 02  5  6  7  8  9
move 03  10 11 12 13 14 15
move 04  16 17 18 19 20 21
move 05  22 23 24 25 26 27
move 06  28 29 30 31 32 33
move 07  34 35 36 37 38 39
move 08  40 41 42 43 44 45
move 09  46 47 48 49 50 51
move 10  52 53 54 55 56 57
move 11  58 59 60 61 62 63
move 12  64 65 66 67 68 69
move 13  70 71 72 73
mv Unknown/0073.05 Unknown/0073.06 13
move 14  74 75 76 77 # potentially missing content?
move 15  78 79 80 81 82 83
move 16  84 85 86 87 88 89
move 17  90 91 92 93 94 95
move 18  96 97 98 99 100 101
move 19  102 103 104 105 106 107
move 20  108 109 110 111 112 113
move 21  114 115 116 117 118 119
move 22  120 121 122 123 124 125
move 23  126 127 128 129 130 131
move 24  132 133 134 135 136 137
move 25  138 139 140 141 142 143
move 26  144 145 146 147 148 149
move 27  150 151 152 153 154 155
