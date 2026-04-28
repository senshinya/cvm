package preprocessor

type scanByte struct {
	b             byte
	offset        int
	needsCleaning bool
}

func scanFile(sm *SourceManager, fileID int, source string, opts Options) ([]PPToken, error) {
	opts = normalizeOptions(opts)
	cleaned := cleanTrigraphs(source, opts)
	var tokens []PPToken
	startOfLine := true
	leadingSpace := false
	pendingClean := false
	for i := 0; i < len(cleaned); {
		ch := cleaned[i]
		if ch.b == '\\' && i+1 < len(cleaned) && cleaned[i+1].b == '\n' {
			// C99 翻译阶段会在词法识别前删除反斜杠换行；这里保留一个空白痕迹给后续 token。
			leadingSpace = true
			pendingClean = true
			i += 2
			continue
		}
		switch ch.b {
		case ' ', '\t', '\r', '\v', '\f':
			leadingSpace = true
			i++
			continue
		case '\n':
			tokens = append(tokens, PPToken{
				Kind:          PPNewline,
				Lexeme:        "\n",
				Location:      sm.Location(fileID, ch.offset),
				NeedsCleaning: ch.needsCleaning || pendingClean,
			})
			startOfLine = true
			leadingSpace = false
			pendingClean = false
			i++
			continue
		}
		if ch.b == '/' && i+1 < len(cleaned) {
			if cleaned[i+1].b == '*' {
				// 注释在预处理阶段被替换成一个空格；块注释内部的换行仍要保留行结构。
				hadNewline := false
				needsClean := ch.needsCleaning || cleaned[i+1].needsCleaning
				i += 2
				for i < len(cleaned) && !(cleaned[i].b == '*' && i+1 < len(cleaned) && cleaned[i+1].b == '/') {
					needsClean = needsClean || cleaned[i].needsCleaning
					if cleaned[i].b == '\n' {
						hadNewline = true
						tokens = append(tokens, PPToken{Kind: PPNewline, Lexeme: "\n", Location: sm.Location(fileID, cleaned[i].offset), NeedsCleaning: needsClean})
						startOfLine = true
						leadingSpace = false
						needsClean = false
					}
					i++
				}
				if i+1 >= len(cleaned) {
					return nil, ppError(sm.Location(fileID, ch.offset), "unterminated comment")
				}
				needsClean = needsClean || cleaned[i].needsCleaning || cleaned[i+1].needsCleaning
				i += 2
				if !hadNewline {
					leadingSpace = true
				}
				pendingClean = pendingClean || needsClean
				continue
			}
			if cleaned[i+1].b == '/' {
				// 行注释同样替换为空格，但终止换行由主循环正常发出。
				pendingClean = pendingClean || ch.needsCleaning || cleaned[i+1].needsCleaning
				i += 2
				for i < len(cleaned) && cleaned[i].b != '\n' {
					pendingClean = pendingClean || cleaned[i].needsCleaning
					i++
				}
				leadingSpace = true
				continue
			}
		}
		if ch.b == '\\' && i+1 < len(cleaned) && (cleaned[i+1].b == 'u' || cleaned[i+1].b == 'U') {
			return nil, ppError(sm.Location(fileID, ch.offset), "unsupported universal-character-name")
		}
		switch {
		case isIdentStart(ch.b):
			start := i
			needsClean := pendingClean
			for i < len(cleaned) && isIdentContinue(cleaned[i].b) {
				needsClean = needsClean || cleaned[i].needsCleaning
				i++
			}
			tokens = append(tokens, makePPToken(PPIdentifier, cleaned[start:i], sm, fileID, startOfLine, leadingSpace, needsClean))
		case isDigit(ch.b) || (ch.b == '.' && i+1 < len(cleaned) && isDigit(cleaned[i+1].b)):
			start := i
			needsClean := pendingClean
			for i < len(cleaned) && isPPNumberByte(cleaned[i].b) {
				needsClean = needsClean || cleaned[i].needsCleaning
				i++
			}
			tokens = append(tokens, makePPToken(PPNumber, cleaned[start:i], sm, fileID, startOfLine, leadingSpace, needsClean))
		case ch.b == '"' || ch.b == '\'':
			kind := PPString
			if ch.b == '\'' {
				kind = PPCharacter
			}
			start := i
			quote := ch.b
			needsClean := pendingClean || ch.needsCleaning
			i++
			for i < len(cleaned) {
				needsClean = needsClean || cleaned[i].needsCleaning
				if cleaned[i].b == '\\' {
					i += 2
					continue
				}
				if cleaned[i].b == quote {
					i++
					break
				}
				if cleaned[i].b == '\n' {
					return nil, ppError(sm.Location(fileID, ch.offset), "unterminated literal")
				}
				i++
			}
			if i > len(cleaned) || cleaned[i-1].b != quote {
				return nil, ppError(sm.Location(fileID, ch.offset), "unterminated literal")
			}
			tokens = append(tokens, makePPToken(kind, cleaned[start:i], sm, fileID, startOfLine, leadingSpace, needsClean))
		default:
			start := i
			lexemeLen := punctuatorLen(cleaned, i)
			needsClean := pendingClean
			for j := 0; j < lexemeLen; j++ {
				needsClean = needsClean || cleaned[i+j].needsCleaning
			}
			i += lexemeLen
			tokens = append(tokens, makePPToken(PPPunctuator, cleaned[start:i], sm, fileID, startOfLine, leadingSpace, needsClean))
		}
		startOfLine = false
		leadingSpace = false
		pendingClean = false
	}
	return tokens, nil
}

func cleanTrigraphs(source string, opts Options) []scanByte {
	out := make([]scanByte, 0, len(source))
	for i := 0; i < len(source); {
		if opts.Std == StandardC99 && i+2 < len(source) {
			if b, ok := translateTrigraph(source[i], source[i+1], source[i+2]); ok {
				out = append(out, scanByte{b: b, offset: i, needsCleaning: true})
				i += 3
				continue
			}
		}
		out = append(out, scanByte{b: source[i], offset: i})
		i++
	}
	return out
}

func makePPToken(kind PPTokenKind, bytes []scanByte, sm *SourceManager, fileID int, startOfLine, leadingSpace, needsClean bool) PPToken {
	buf := make([]byte, len(bytes))
	for i, b := range bytes {
		buf[i] = b.b
	}
	return PPToken{
		Kind:          kind,
		Lexeme:        string(buf),
		Location:      sm.Location(fileID, bytes[0].offset),
		StartOfLine:   startOfLine,
		LeadingSpace:  leadingSpace,
		NeedsCleaning: needsClean,
	}
}

func punctuatorLen(bytes []scanByte, i int) int {
	if i+2 < len(bytes) && bytes[i].b == '.' && bytes[i+1].b == '.' && bytes[i+2].b == '.' {
		return 3
	}
	if i+2 < len(bytes) {
		three := string([]byte{bytes[i].b, bytes[i+1].b, bytes[i+2].b})
		if three == "<<=" || three == ">>=" {
			return 3
		}
	}
	if i+1 < len(bytes) {
		two := string([]byte{bytes[i].b, bytes[i+1].b})
		switch two {
		case "##", "->", "++", "--", "<<", ">>", "<=", ">=", "==", "!=", "&&", "||",
			"*=", "/=", "%=", "+=", "-=", "&=", "^=", "|=":
			return 2
		}
	}
	return 1
}

func translateTrigraph(a, b, c byte) (byte, bool) {
	if a != '?' || b != '?' {
		return 0, false
	}
	switch c {
	case '=':
		return '#', true
	case '/':
		return '\\', true
	case '\'':
		return '^', true
	case '(':
		return '[', true
	case ')':
		return ']', true
	case '!':
		return '|', true
	case '<':
		return '{', true
	case '>':
		return '}', true
	case '-':
		return '~', true
	default:
		return 0, false
	}
}

func isIdentStart(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isIdentContinue(b byte) bool {
	return isIdentStart(b) || isDigit(b)
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isPPNumberByte(b byte) bool {
	return isIdentContinue(b) || b == '.' || b == '+' || b == '-'
}
