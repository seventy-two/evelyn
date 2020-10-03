package dota

func getShortHero(id int) (out []string) {
	switch id {
	case 1:
		return []string{"Anti", "Mage"}
	case 2:
		return []string{"Axe"}
	case 3:
		return []string{"Bane"}
	case 4:
		return []string{"Blood", "Seeker"}
	case 5:
		return []string{"Crystal", "Maiden"}
	case 6:
		return []string{"Drow", "Ranger"}
	case 7:
		return []string{"Earth", "Shaker"}
	case 8:
		return []string{"Juggernaut"}
	case 9:
		return []string{"Mirana"}
	case 11:
		return []string{"Shadow", "Fiend"}
	case 10:
		return []string{"Morphling"}
	case 12:
		return []string{"Phantom", "Lancer"}
	case 13:
		return []string{"Puck"}
	case 14:
		return []string{"Pudge"}
	case 15:
		return []string{"Razor"}
	case 16:
		return []string{"Sand", "King"}
	case 17:
		return []string{"Storm", "Spirit"}
	case 18:
		return []string{"Sven"}
	case 19:
		return []string{"Tiny"}
	case 20:
		return []string{"Vengeful", "Spirit"}
	case 21:
		return []string{"Wind", "Ranger"}
	case 22:
		return []string{"Zeus"}
	case 23:
		return []string{"Kunkka"}
	case 25:
		return []string{"Lina"}
	case 31:
		return []string{"Lich"}
	case 26:
		return []string{"Lion"}
	case 27:
		return []string{"Shadow", "Shaman"}
	case 28:
		return []string{"Slardar"}
	case 29:
		return []string{"Tide", "Hunter"}
	case 30:
		return []string{"Witch", "Doctor"}
	case 32:
		return []string{"Riki"}
	case 33:
		return []string{"Enigma"}
	case 34:
		return []string{"Tinker"}
	case 35:
		return []string{"Sniper"}
	case 36:
		return []string{"Necrophos"}
	case 37:
		return []string{"Warlock"}
	case 38:
		return []string{"Beast", "Master"}
	case 39:
		return []string{"Queen", "of Pain"}
	case 40:
		return []string{"Venomancer"}
	case 41:
		return []string{"Faceless", "Void"}
	case 42:
		return []string{"Wraith", "King"}
	case 43:
		return []string{"Death", "Prophet"}
	case 44:
		return []string{"Phantom", "Assassin"}
	case 45:
		return []string{"Pugna"}
	case 46:
		return []string{"Templar", "Assassin"}
	case 47:
		return []string{"Viper"}
	case 48:
		return []string{"Luna"}
	case 49:
		return []string{"Dragon", "Knight"}
	case 50:
		return []string{"Dazzle"}
	case 51:
		return []string{"Clockwerk"}
	case 52:
		return []string{"Leshrac"}
	case 53:
		return []string{"Natures", "Prophet"}
	case 54:
		return []string{"Life", "stealer"}
	case 55:
		return []string{"Dark", "Seer"}
	case 56:
		return []string{"Clinkz"}
	case 57:
		return []string{"Omni", "knight"}
	case 58:
		return []string{"Enchantress"}
	case 59:
		return []string{"Huskar"}
	case 60:
		return []string{"Night", "Stalker"}
	case 61:
		return []string{"Brood", "Mother"}
	case 62:
		return []string{"Bounty", "Hunter"}
	case 63:
		return []string{"Weaver"}
	case 64:
		return []string{"Jakiro"}
	case 65:
		return []string{"Bat", "Rider"}
	case 66:
		return []string{"Chen"}
	case 67:
		return []string{"Spectre"}
	case 69:
		return []string{"Doom"}
	case 68:
		return []string{"Ancient", "Apparition"}
	case 70:
		return []string{"Ursa"}
	case 71:
		return []string{"Spirit", "Breaker"}
	case 72:
		return []string{"Gyro", "copter"}
	case 73:
		return []string{"Alchemist"}
	case 74:
		return []string{"Invoker"}
	case 75:
		return []string{"Silencer"}
	case 76:
		return []string{"Outworld", "Devourer"}
	case 77:
		return []string{"Lycan"}
	case 78:
		return []string{"Brew"}
	case 79:
		return []string{"Shadow", "Demon"}
	case 80:
		return []string{"Lone", "Druid"}
	case 81:
		return []string{"Chaos", "Knight"}
	case 82:
		return []string{"Meepo"}
	case 83:
		return []string{"Treant", "Protector"}
	case 84:
		return []string{"Ogre", "Magi"}
	case 85:
		return []string{"Undying"}
	case 86:
		return []string{"Rubick"}
	case 87:
		return []string{"Disruptor"}
	case 88:
		return []string{"Nyx", "Assassin"}
	case 89:
		return []string{"Naga", "Siren"}
	case 90:
		return []string{"Keeper", "of the Light"}
	case 91:
		return []string{"Io"}
	case 92:
		return []string{"Visage"}
	case 93:
		return []string{"Slark"}
	case 94:
		return []string{"Medusa"}
	case 95:
		return []string{"Troll", "Warlord"}
	case 96:
		return []string{"Centaur", "Warrunner"}
	case 97:
		return []string{"Magnus"}
	case 98:
		return []string{"Timbersaw"}
	case 99:
		return []string{"Bristle", "back"}
	case 100:
		return []string{"Tusk"}
	case 101:
		return []string{"Skywrath", "Mage"}
	case 102:
		return []string{"Abaddon"}
	case 103:
		return []string{"Elder", "Titan"}
	case 104:
		return []string{"Legion", "Commander"}
	case 106:
		return []string{"Ember", "Spirit"}
	case 107:
		return []string{"Earth", "Spirit"}
	case 109:
		return []string{"Terror", "blade"}
	case 110:
		return []string{"Phoenix"}
	case 111:
		return []string{"Oracle"}
	case 105:
		return []string{"Techies"}
	case 112:
		return []string{"Winter", "Wyvern"}
	case 113:
		return []string{"Arc", "Warden"}
	case 108:
		return []string{"Abyssal", "Underlord"}
	case 114:
		return []string{"Monkey", "King"}
	case 120:
		return []string{"Pangolier"}
	case 119:
		return []string{"Dark", "Willow"}
	default:
		return nil
	}
}
