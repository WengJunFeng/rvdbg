//-----------------------------------------------------------------------------
/*

Decode 32-bit JTAG IDCODE numbers.

*/
//-----------------------------------------------------------------------------

package jtag

import (
	"fmt"
	"strings"

	"github.com/deadsy/rvdbg/util"
)

//-----------------------------------------------------------------------------

// mfgName maps a manufacturer id to name
var mfgName = map[uint]string{
	0x001: "AMD",
	0x002: "AMI",
	0x003: "Fairchild",
	0x004: "Fujitsu",
	0x005: "GTE",
	0x006: "Harris",
	0x007: "Hitachi",
	0x008: "Inmos",
	0x009: "Intel",
	0x00a: "I.T.T.",
	0x00b: "Intersil",
	0x00c: "Monolithic Memories",
	0x00d: "Mostek",
	0x00e: "Freescale (Motorola)",
	0x00f: "National",
	0x010: "NEC",
	0x011: "RCA",
	0x012: "Raytheon",
	0x013: "Conexant (Rockwell)",
	0x014: "Seeq",
	0x015: "NXP (Philips)",
	0x016: "Synertek",
	0x017: "Texas Instruments",
	0x018: "Toshiba",
	0x019: "Xicor",
	0x01a: "Zilog",
	0x01b: "Eurotechnique",
	0x01c: "Mitsubishi",
	0x01d: "Lucent (AT&T)",
	0x01e: "Exel",
	0x01f: "Atmel",
	0x020: "SGS/Thomson",
	0x021: "Lattice Semi.",
	0x022: "NCR",
	0x023: "Wafer Scale Integration",
	0x024: "IBM",
	0x025: "Tristar",
	0x026: "Visic",
	0x027: "Intl. CMOS Technology",
	0x028: "SSSI",
	0x029: "MicrochipTechnology",
	0x02a: "Ricoh Ltd.",
	0x02b: "VLSI",
	0x02c: "Micron Technology",
	0x02d: "Hynix Semiconductor (Hyundai Electronics)",
	0x02e: "OKI Semiconductor",
	0x02f: "ACTEL",
	0x030: "Sharp",
	0x031: "Catalyst",
	0x032: "Panasonic",
	0x033: "IDT",
	0x034: "Cypress",
	0x035: "DEC",
	0x036: "LSI Logic",
	0x037: "Zarlink (Plessey)",
	0x038: "UTMC",
	0x039: "Thinking Machine",
	0x03a: "Thomson CSF",
	0x03b: "Integrated CMOS (Vertex)",
	0x03c: "Honeywell",
	0x03d: "Tektronix",
	0x03e: "Oracle Corporation",
	0x03f: "Silicon Storage Technology",
	0x040: "ProMos/Mosel Vitelic",
	0x041: "Infineon (Siemens)",
	0x042: "Macronix",
	0x043: "Xerox",
	0x044: "Plus Logic",
	0x045: "SanDisk Corporation",
	0x046: "Elan Circuit Tech.",
	0x047: "European Silicon Str.",
	0x048: "Apple Computer",
	0x049: "Xilinx",
	0x04a: "Compaq",
	0x04b: "Protocol Engines",
	0x04c: "SCI",
	0x04d: "Seiko Instruments",
	0x04e: "Samsung",
	0x04f: "I3 Design System",
	0x050: "Klic",
	0x051: "Crosspoint Solutions",
	0x052: "Alliance Semiconductor",
	0x053: "Tandem",
	0x054: "Hewlett-Packard",
	0x055: "Integrated Silicon Solutions",
	0x056: "Brooktree",
	0x057: "New Media",
	0x058: "MHS Electronic",
	0x059: "Performance Semi.",
	0x05a: "Winbond Electronic",
	0x05b: "Kawasaki Steel",
	0x05c: "Bright Micro",
	0x05d: "TECMAR",
	0x05e: "Exar",
	0x05f: "PCMCIA",
	0x060: "LG Semi (Goldstar)",
	0x061: "Northern Telecom",
	0x062: "Sanyo",
	0x063: "Array Microsystems",
	0x064: "Crystal Semiconductor",
	0x065: "Analog Devices",
	0x066: "PMC-Sierra",
	0x067: "Asparix",
	0x068: "Convex Computer",
	0x069: "Quality Semiconductor",
	0x06a: "Nimbus Technology",
	0x06b: "Transwitch",
	0x06c: "Micronas (ITT Intermetall)",
	0x06d: "Cannon",
	0x06e: "Altera",
	0x06f: "NEXCOM",
	0x070: "QUALCOMM",
	0x071: "Sony",
	0x072: "Cray Research",
	0x073: "AMS(Austria Micro)",
	0x074: "Vitesse",
	0x075: "Aster Electronics",
	0x076: "Bay Networks (Synoptic)",
	0x077: "Zentrum/ZMD",
	0x078: "TRW",
	0x079: "Thesys",
	0x07a: "Solbourne Computer",
	0x07b: "Allied-Signal",
	0x07c: "Dialog Semiconductor",
	0x07d: "Media Vision",
	0x07e: "Numonyx Corporation",
	0x081: "Cirrus Logic",
	0x082: "National Instruments",
	0x083: "ILC Data Device",
	0x084: "Alcatel Mietec",
	0x085: "Micro Linear",
	0x086: "Univ. of NC",
	0x087: "JTAG Technologies",
	0x088: "BAE Systems (Loral)",
	0x089: "Nchip",
	0x08a: "Galileo Tech",
	0x08b: "Bestlink Systems",
	0x08c: "Graychip",
	0x08d: "GENNUM",
	0x08e: "VideoLogic",
	0x08f: "Robert Bosch",
	0x090: "Chip Express",
	0x091: "DATARAM",
	0x092: "United Microelectronics Corp.",
	0x093: "TCSI",
	0x094: "Smart Modular",
	0x095: "Hughes Aircraft",
	0x096: "Lanstar Semiconductor",
	0x097: "Qlogic",
	0x098: "Kingston",
	0x099: "Music Semi",
	0x09a: "Ericsson Components",
	0x09b: "SpaSE",
	0x09c: "Eon Silicon Devices",
	0x09d: "Programmable Micro Corp",
	0x09e: "DoD",
	0x09f: "Integ. Memories Tech.",
	0x0a0: "Corollary Inc.",
	0x0a1: "Dallas Semiconductor",
	0x0a2: "Omnivision",
	0x0a3: "EIV(Switzerland)",
	0x0a4: "Novatel Wireless",
	0x0a5: "Zarlink (Mitel)",
	0x0a6: "Clearpoint",
	0x0a7: "Cabletron",
	0x0a8: "STEC (Silicon Tech)",
	0x0a9: "Vanguard",
	0x0aa: "Hagiwara Sys-Com",
	0x0ab: "Vantis",
	0x0ac: "Celestica",
	0x0ad: "Century",
	0x0ae: "Hal Computers",
	0x0af: "Rohm Company Ltd.",
	0x0b0: "Juniper Networks",
	0x0b1: "Libit Signal Processing",
	0x0b2: "Mushkin Enhanced Memory",
	0x0b3: "Tundra Semiconductor",
	0x0b4: "Adaptec Inc.",
	0x0b5: "LightSpeed Semi.",
	0x0b6: "ZSP Corp.",
	0x0b7: "AMIC Technology",
	0x0b8: "Adobe Systems",
	0x0b9: "Dynachip",
	0x0ba: "PNY Electronics",
	0x0bb: "Newport Digital",
	0x0bc: "MMC Networks",
	0x0bd: "T Square",
	0x0be: "Seiko Epson",
	0x0bf: "Broadcom",
	0x0c0: "Viking Components",
	0x0c1: "V3 Semiconductor",
	0x0c2: "Flextronics (Orbit Semiconductor)",
	0x0c3: "Suwa Electronics",
	0x0c4: "Transmeta",
	0x0c5: "Micron CMS",
	0x0c6: "American Computer & Digital Components Inc",
	0x0c7: "Enhance 3000 Inc",
	0x0c8: "Tower Semiconductor",
	0x0c9: "CPU Design",
	0x0ca: "Price Point",
	0x0cb: "Maxim Integrated Product",
	0x0cc: "Tellabs",
	0x0cd: "Centaur Technology",
	0x0ce: "Unigen Corporation",
	0x0cf: "Transcend Information",
	0x0d0: "Memory Card Technology",
	0x0d1: "CKD Corporation Ltd.",
	0x0d2: "Capital Instruments, Inc.",
	0x0d3: "Aica Kogyo, Ltd.",
	0x0d4: "Linvex Technology",
	0x0d5: "MSC Vertriebs GmbH",
	0x0d6: "AKM Company, Ltd.",
	0x0d7: "Dynamem, Inc.",
	0x0d8: "NERA ASA",
	0x0d9: "GSI Technology",
	0x0da: "Dane-Elec (C Memory)",
	0x0db: "Acorn Computers",
	0x0dc: "Lara Technology",
	0x0dd: "Oak Technology, Inc.",
	0x0de: "Itec Memory",
	0x0df: "Tanisys Technology",
	0x0e0: "Truevision",
	0x0e1: "Wintec Industries",
	0x0e2: "Super PC Memory",
	0x0e3: "MGV Memory",
	0x0e4: "Galvantech",
	0x0e5: "Gadzoox Networks",
	0x0e6: "Multi Dimensional Cons.",
	0x0e7: "GateField",
	0x0e8: "Integrated Memory System",
	0x0e9: "Triscend",
	0x0ea: "XaQti",
	0x0eb: "Goldenram",
	0x0ec: "Clear Logic",
	0x0ed: "Cimaron Communications",
	0x0ee: "Nippon Steel Semi. Corp.",
	0x0ef: "Advantage Memory",
	0x0f0: "AMCC",
	0x0f1: "LeCroy",
	0x0f2: "Yamaha Corporation",
	0x0f3: "Digital Microwave",
	0x0f4: "NetLogic Microsystems",
	0x0f5: "MIMOS Semiconductor",
	0x0f6: "Advanced Fibre",
	0x0f7: "BF Goodrich Data.",
	0x0f8: "Epigram",
	0x0f9: "Acbel Polytech Inc.",
	0x0fa: "Apacer Technology",
	0x0fb: "Admor Memory",
	0x0fc: "FOXCONN",
	0x0fd: "Quadratics Superconductor",
	0x0fe: "3COM",
	0x101: "Camintonn Corporation",
	0x102: "ISOA Incorporated",
	0x103: "Agate Semiconductor",
	0x104: "ADMtek Incorporated",
	0x105: "HYPERTEC",
	0x106: "Adhoc Technologies",
	0x107: "MOSAID Technologies",
	0x108: "Ardent Technologies",
	0x109: "Switchcore",
	0x10a: "Cisco Systems, Inc.",
	0x10b: "Allayer Technologies",
	0x10c: "WorkX AG (Wichman)",
	0x10d: "Oasis Semiconductor",
	0x10e: "Novanet Semiconductor",
	0x10f: "E-M Solutions",
	0x110: "Power General",
	0x111: "Advanced Hardware Arch.",
	0x112: "Inova Semiconductors GmbH",
	0x113: "Telocity",
	0x114: "Delkin Devices",
	0x115: "Symagery Microsystems",
	0x116: "C-Port Corporation",
	0x117: "SiberCore Technologies",
	0x118: "Southland Microsystems",
	0x119: "Malleable Technologies",
	0x11a: "Kendin Communications",
	0x11b: "Great Technology Microcomputer",
	0x11c: "Sanmina Corporation",
	0x11d: "HADCO Corporation",
	0x11e: "Corsair",
	0x11f: "Actrans System Inc.",
	0x120: "ALPHA Technologies",
	0x121: "Silicon Laboratories, Inc. (Cygnal)",
	0x122: "Artesyn Technologies",
	0x123: "Align Manufacturing",
	0x124: "Peregrine Semiconductor",
	0x125: "Chameleon Systems",
	0x126: "Aplus Flash Technology",
	0x127: "MIPS Technologies",
	0x128: "Chrysalis ITS",
	0x129: "ADTEC Corporation",
	0x12a: "Kentron Technologies",
	0x12b: "Win Technologies",
	0x12c: "Tachyon Semiconductor (ASIC)",
	0x12d: "Extreme Packet Devices",
	0x12e: "RF Micro Devices",
	0x12f: "Siemens AG",
	0x130: "Sarnoff Corporation",
	0x131: "Itautec SA",
	0x132: "Radiata Inc.",
	0x133: "Benchmark Elect. (AVEX)",
	0x134: "Legend",
	0x135: "SpecTek Incorporated",
	0x136: "Hi/fn",
	0x137: "Enikia Incorporated",
	0x138: "SwitchOn Networks",
	0x139: "AANetcom Incorporated",
	0x13a: "Micro Memory Bank",
	0x13b: "ESS Technology",
	0x13c: "Virata Corporation",
	0x13d: "Excess Bandwidth",
	0x13e: "West Bay Semiconductor",
	0x13f: "DSP Group",
	0x140: "Newport Communications",
	0x141: "Chip2Chip Incorporated",
	0x142: "Phobos Corporation",
	0x143: "Intellitech Corporation",
	0x144: "Nordic VLSI ASA",
	0x145: "Ishoni Networks",
	0x146: "Silicon Spice",
	0x147: "Alchemy Semiconductor",
	0x148: "Agilent Technologies",
	0x149: "Centillium Communications",
	0x14a: "W.L. Gore",
	0x14b: "HanBit Electronics",
	0x14c: "GlobeSpan",
	0x14d: "Element 14",
	0x14e: "Pycon",
	0x14f: "Saifun Semiconductors",
	0x150: "Sibyte, Incorporated",
	0x151: "MetaLink Technologies",
	0x152: "Feiya Technology",
	0x153: "I & C Technology",
	0x154: "Shikatronics",
	0x155: "Elektrobit",
	0x156: "Megic",
	0x157: "Com-Tier",
	0x158: "Malaysia Micro Solutions",
	0x159: "Hyperchip",
	0x15a: "Gemstone Communications",
	0x15b: "Anadigm (Anadyne)",
	0x15c: "3ParData",
	0x15d: "Mellanox Technologies",
	0x15e: "Tenx Technologies",
	0x15f: "Helix AG",
	0x160: "Domosys",
	0x161: "Skyup Technology",
	0x162: "HiNT Corporation",
	0x163: "Chiaro",
	0x164: "MDT Technologies GmbH",
	0x165: "Exbit Technology A/S",
	0x166: "Integrated Technology Express",
	0x167: "AVED Memory",
	0x168: "Legerity",
	0x169: "Jasmine Networks",
	0x16a: "Caspian Networks",
	0x16b: "nCUBE",
	0x16c: "Silicon Access Networks",
	0x16d: "FDK Corporation",
	0x16e: "High Bandwidth Access",
	0x16f: "MultiLink Technology",
	0x170: "BRECIS",
	0x171: "World Wide Packets",
	0x172: "APW",
	0x173: "Chicory Systems",
	0x174: "Xstream Logic",
	0x175: "Fast-Chip",
	0x176: "Zucotto Wireless",
	0x177: "Realchip",
	0x178: "Galaxy Power",
	0x179: "eSilicon",
	0x17a: "Morphics Technology",
	0x17b: "Accelerant Networks",
	0x17c: "Silicon Wave",
	0x17d: "SandCraft",
	0x17e: "Elpida",
	0x181: "Solectron",
	0x182: "Optosys Technologies",
	0x183: "Buffalo (Formerly Melco)",
	0x184: "TriMedia Technologies",
	0x185: "Cyan Technologies",
	0x186: "Global Locate",
	0x187: "Optillion",
	0x188: "Terago Communications",
	0x189: "Ikanos Communications",
	0x18a: "Princeton Technology",
	0x18b: "Nanya Technology",
	0x18c: "Elite Flash Storage",
	0x18d: "Mysticom",
	0x18e: "LightSand Communications",
	0x18f: "ATI Technologies",
	0x190: "Agere Systems",
	0x191: "NeoMagic",
	0x192: "AuroraNetics",
	0x193: "Golden Empire",
	0x194: "Mushkin",
	0x195: "Tioga Technologies",
	0x196: "Netlist",
	0x197: "TeraLogic",
	0x198: "Cicada Semiconductor",
	0x199: "Centon Electronics",
	0x19a: "Tyco Electronics",
	0x19b: "Magis Works",
	0x19c: "Zettacom",
	0x19d: "Cogency Semiconductor",
	0x19e: "Chipcon AS",
	0x19f: "Aspex Technology",
	0x1a0: "F5 Networks",
	0x1a1: "Programmable Silicon Solutions",
	0x1a2: "ChipWrights",
	0x1a3: "Acorn Networks",
	0x1a4: "Quicklogic",
	0x1a5: "Kingmax Semiconductor",
	0x1a6: "BOPS",
	0x1a7: "Flasys",
	0x1a8: "BitBlitz Communications",
	0x1a9: "eMemory Technology",
	0x1aa: "Procket Networks",
	0x1ab: "Purple Ray",
	0x1ac: "Trebia Networks",
	0x1ad: "Delta Electronics",
	0x1ae: "Onex Communications",
	0x1af: "Ample Communications",
	0x1b0: "Memory Experts Intl",
	0x1b1: "Astute Networks",
	0x1b2: "Azanda Network Devices",
	0x1b3: "Dibcom",
	0x1b4: "Tekmos",
	0x1b5: "API NetWorks",
	0x1b6: "Bay Microsystems",
	0x1b7: "Firecron Ltd",
	0x1b8: "Resonext Communications",
	0x1b9: "Tachys Technologies",
	0x1ba: "Equator Technology",
	0x1bb: "Concept Computer",
	0x1bc: "SILCOM",
	0x1bd: "3Dlabs",
	0x1be: "c?t Magazine",
	0x1bf: "Sanera Systems",
	0x1c0: "Silicon Packets",
	0x1c1: "Viasystems Group",
	0x1c2: "Simtek",
	0x1c3: "Semicon Devices Singapore",
	0x1c4: "Satron Handelsges",
	0x1c5: "Improv Systems",
	0x1c6: "INDUSYS GmbH",
	0x1c7: "Corrent",
	0x1c8: "Infrant Technologies",
	0x1c9: "Ritek Corp",
	0x1ca: "empowerTel Networks",
	0x1cb: "Hypertec",
	0x1cc: "Cavium Networks",
	0x1cd: "PLX Technology",
	0x1ce: "Massana Design",
	0x1cf: "Intrinsity",
	0x1d0: "Valence Semiconductor",
	0x1d1: "Terawave Communications",
	0x1d2: "IceFyre Semiconductor",
	0x1d3: "Primarion",
	0x1d4: "Picochip Designs Ltd",
	0x1d5: "Silverback Systems",
	0x1d6: "Jade Star Technologies",
	0x1d7: "Pijnenburg Securealink",
	0x1d8: "takeMS International AG",
	0x1d9: "Cambridge Silicon Radio",
	0x1da: "Swissbit",
	0x1db: "Nazomi Communications",
	0x1dc: "eWave System",
	0x1dd: "Rockwell Collins",
	0x1de: "Picocel Co. Ltd. (Paion)",
	0x1df: "Alphamosaic Ltd",
	0x1e0: "Sandburst",
	0x1e1: "SiCon Video",
	0x1e2: "NanoAmp Solutions",
	0x1e3: "Ericsson Technology",
	0x1e4: "PrairieComm",
	0x1e5: "Mitac International",
	0x1e6: "Layer N Networks",
	0x1e7: "MtekVision (Atsana)",
	0x1e8: "Allegro Networks",
	0x1e9: "Marvell Semiconductors",
	0x1ea: "Netergy Microelectronic",
	0x1eb: "NVIDIA",
	0x1ec: "Internet Machines",
	0x1ed: "Peak Electronics",
	0x1ee: "Litchfield Communication",
	0x1ef: "Accton Technology",
	0x1f0: "Teradiant Networks",
	0x1f1: "Scaleo Chip",
	0x1f2: "Cortina Systems",
	0x1f3: "RAM Components",
	0x1f4: "Raqia Networks",
	0x1f5: "ClearSpeed",
	0x1f6: "Matsushita Battery",
	0x1f7: "Xelerated",
	0x1f8: "SimpleTech",
	0x1f9: "Utron Technology",
	0x1fa: "Astec International",
	0x1fb: "AVM gmbH",
	0x1fc: "Redux Communications",
	0x1fd: "Dot Hill Systems",
	0x1fe: "TeraChip",
	0x201: "T-RAM Incorporated",
	0x202: "Innovics Wireless",
	0x203: "Teknovus",
	0x204: "KeyEye Communications",
	0x205: "Runcom Technologies",
	0x206: "RedSwitch",
	0x207: "Dotcast",
	0x208: "Silicon Mountain Memory",
	0x209: "Signia Technologies",
	0x20a: "Pixim",
	0x20b: "Galazar Networks",
	0x20c: "White Electronic Designs",
	0x20d: "Patriot Scientific",
	0x20e: "Neoaxiom Corporation",
	0x20f: "3Y Power Technology",
	0x210: "Scaleo Chip",
	0x211: "Potentia Power Systems",
	0x212: "C-guys Incorporated",
	0x213: "Digital Communications Technology Incorporated",
	0x214: "Silicon-Based Technology",
	0x215: "Fulcrum Microsystems",
	0x216: "Positivo Informatica Ltd",
	0x217: "XIOtech Corporation",
	0x218: "PortalPlayer",
	0x219: "Zhiying Software",
	0x21a: "ParkerVision, Inc.",
	0x21b: "Phonex Broadband",
	0x21c: "Skyworks Solutions",
	0x21d: "Entropic Communications",
	0x21e: "Pacific Force Technology",
	0x21f: "Zensys A/S",
	0x220: "Legend Silicon Corp.",
	0x221: "Sci-worx GmbH",
	0x222: "SMSC (Standard Microsystems)",
	0x223: "Renesas Electronics",
	0x224: "Raza Microelectronics",
	0x225: "Phyworks",
	0x226: "MediaTek",
	0x227: "Non-cents Productions",
	0x228: "US Modular",
	0x229: "Wintegra Ltd.",
	0x22a: "Mathstar",
	0x22b: "StarCore",
	0x22c: "Oplus Technologies",
	0x22d: "Mindspeed",
	0x22e: "Just Young Computer",
	0x22f: "Radia Communications",
	0x230: "OCZ",
	0x231: "Emuzed",
	0x232: "LOGIC Devices",
	0x233: "Inphi Corporation",
	0x234: "Quake Technologies",
	0x235: "Vixel",
	0x236: "SolusTek",
	0x237: "Kongsberg Maritime",
	0x238: "Faraday Technology",
	0x239: "Altium Ltd.",
	0x23a: "Insyte",
	0x23b: "ARM Ltd.",
	0x23c: "DigiVision",
	0x23d: "Vativ Technologies",
	0x23e: "Endicott Interconnect Technologies",
	0x23f: "Pericom",
	0x240: "Bandspeed",
	0x241: "LeWiz Communications",
	0x242: "CPU Technology",
	0x243: "Ramaxel Technology",
	0x244: "DSP Group",
	0x245: "Axis Communications",
	0x246: "Legacy Electronics",
	0x247: "Chrontel",
	0x248: "Powerchip Semiconductor",
	0x249: "MobilEye Technologies",
	0x24a: "Excel Semiconductor",
	0x24b: "A-DATA Technology",
	0x24c: "VirtualDigm",
	0x24d: "G Skill Intl",
	0x24e: "Quanta Computer",
	0x24f: "Yield Microelectronics",
	0x250: "Afa Technologies",
	0x251: "KINGBOX Technology Co. Ltd.",
	0x252: "Ceva",
	0x253: "iStor Networks",
	0x254: "Advance Modules",
	0x255: "Microsoft",
	0x256: "Open-Silicon",
	0x257: "Goal Semiconductor",
	0x258: "ARC International",
	0x259: "Simmtec",
	0x25a: "Metanoia",
	0x25b: "Key Stream",
	0x25c: "Lowrance Electronics",
	0x25d: "Adimos",
	0x25e: "SiGe Semiconductor",
	0x25f: "Fodus Communications",
	0x260: "Credence Systems Corp.",
	0x261: "Genesis Microchip Inc.",
	0x262: "Vihana, Inc.",
	0x263: "WIS Technologies",
	0x264: "GateChange Technologies",
	0x265: "High Density Devices AS",
	0x266: "Synopsys",
	0x267: "Gigaram",
	0x268: "Enigma Semiconductor Inc.",
	0x269: "Century Micro Inc.",
	0x26a: "Icera Semiconductor",
	0x26b: "Mediaworks Integrated Systems",
	0x26c: "O'Neil Product Development",
	0x26d: "Supreme Top Technology Ltd.",
	0x26e: "MicroDisplay Corporation",
	0x26f: "Team Group Inc.",
	0x270: "Sinett Corporation",
	0x271: "Toshiba Corporation",
	0x272: "Tensilica",
	0x273: "SiRF Technology",
	0x274: "Bacoc Inc.",
	0x275: "SMaL Camera Technologies",
	0x276: "Thomson SC",
	0x277: "Airgo Networks",
	0x278: "Wisair Ltd.",
	0x279: "SigmaTel",
	0x27a: "Arkados",
	0x27b: "Compete IT gmbH Co. KG",
	0x27c: "Eudar Technology Inc.",
	0x27d: "Focus Enhancements",
	0x27e: "Xyratex",
	0x281: "Specular Networks",
	0x282: "Patriot Memory (PDP Systems)",
	0x283: "U-Chip Technology Corp.",
	0x284: "Silicon Optix",
	0x285: "Greenfield Networks",
	0x286: "CompuRAM GmbH",
	0x287: "Stargen, Inc.",
	0x288: "NetCell Corporation",
	0x289: "Excalibrus Technologies Ltd",
	0x28a: "SCM Microsystems",
	0x28b: "Xsigo Systems, Inc.",
	0x28c: "CHIPS & Systems Inc",
	0x28d: "Tier 1 Multichip Solutions",
	0x28e: "CWRL Labs",
	0x28f: "Teradici",
	0x290: "Gigaram, Inc.",
	0x291: "g2 Microsystems",
	0x292: "PowerFlash Semiconductor",
	0x293: "P.A. Semi, Inc.",
	0x294: "NovaTech Solutions, S.A.",
	0x295: "c2 Microsystems, Inc.",
	0x296: "Level5 Networks",
	0x297: "COS Memory AG",
	0x298: "Innovasic Semiconductor",
	0x299: "02IC Co. Ltd",
	0x29a: "Tabula, Inc.",
	0x29b: "Crucial Technology",
	0x29c: "Chelsio Communications",
	0x29d: "Solarflare Communications",
	0x29e: "Xambala Inc.",
	0x29f: "EADS Astrium",
	0x2a0: "Terra Semiconductor, Inc.",
	0x2a1: "Imaging Works, Inc.",
	0x2a2: "Astute Networks, Inc.",
	0x2a3: "Tzero",
	0x2a4: "Emulex",
	0x2a5: "Power-One",
	0x2a6: "Pulse~LINK Inc.",
	0x2a7: "Hon Hai Precision Industry",
	0x2a8: "White Rock Networks Inc.",
	0x2a9: "Telegent Systems USA, Inc.",
	0x2aa: "Atrua Technologies, Inc.",
	0x2ab: "Acbel Polytech Inc.",
	0x2ac: "eRide Inc.",
	0x2ad: "ULi Electronics Inc.",
	0x2ae: "Magnum Semiconductor Inc.",
	0x2af: "neoOne Technology, Inc.",
	0x2b0: "Connex Technology, Inc.",
	0x2b1: "Stream Processors, Inc.",
	0x2b2: "Focus Enhancements",
	0x2b3: "Telecis Wireless, Inc.",
	0x2b4: "uNav Microelectronics",
	0x2b5: "Tarari, Inc.",
	0x2b6: "Ambric, Inc.",
	0x2b7: "Newport Media, Inc.",
	0x2b8: "VMTS",
	0x2b9: "Enuclia Semiconductor, Inc.",
	0x2ba: "Virtium Technology Inc.",
	0x2bb: "Solid State System Co., Ltd.",
	0x2bc: "Kian Tech LLC",
	0x2bd: "Artimi",
	0x2be: "Power Quotient International",
	0x2bf: "Avago Technologies",
	0x2c0: "ADTechnology",
	0x2c1: "Sigma Designs",
	0x2c2: "SiCortex, Inc.",
	0x2c3: "Ventura Technology Group",
	0x2c4: "eASIC",
	0x2c5: "M.H.S. SAS",
	0x2c6: "Micro Star International",
	0x2c7: "Rapport Inc.",
	0x2c8: "Makway International",
	0x2c9: "Broad Reach Engineering Co.",
	0x2ca: "Semiconductor Mfg Intl Corp",
	0x2cb: "SiConnect",
	0x2cc: "FCI USA Inc.",
	0x2cd: "Validity Sensors",
	0x2ce: "Coney Technology Co. Ltd.",
	0x2cf: "Spans Logic",
	0x2d0: "Neterion Inc.",
	0x2d1: "Qimonda",
	0x2d2: "New Japan Radio Co. Ltd.",
	0x2d3: "Velogix",
	0x2d4: "Montalvo Systems",
	0x2d5: "iVivity Inc.",
	0x2d6: "Walton Chaintech",
	0x2d7: "AENEON",
	0x2d8: "Lorom Industrial Co. Ltd.",
	0x2d9: "Radiospire Networks",
	0x2da: "Sensio Technologies, Inc.",
	0x2db: "Nethra Imaging",
	0x2dc: "Hexon Technology Pte Ltd",
	0x2dd: "CompuStocx (CSX)",
	0x2de: "Methode Electronics, Inc.",
	0x2df: "Connect One Ltd.",
	0x2e0: "Opulan Technologies",
	0x2e1: "Septentrio NV",
	0x2e2: "Goldenmars Technology Inc.",
	0x2e3: "Kreton Corporation",
	0x2e4: "Cochlear Ltd.",
	0x2e5: "Altair Semiconductor",
	0x2e6: "NetEffect, Inc.",
	0x2e7: "Spansion, Inc.",
	0x2e8: "Taiwan Semiconductor Mfg",
	0x2e9: "Emphany Systems Inc.",
	0x2ea: "ApaceWave Technologies",
	0x2eb: "Mobilygen Corporation",
	0x2ec: "Tego",
	0x2ed: "Cswitch Corporation",
	0x2ee: "Haier (Beijing) IC Design Co.",
	0x2ef: "MetaRAM",
	0x2f0: "Axel Electronics Co. Ltd.",
	0x2f1: "Tilera Corporation",
	0x2f2: "Aquantia",
	0x2f3: "Vivace Semiconductor",
	0x2f4: "Redpine Signals",
	0x2f5: "Octalica",
	0x2f6: "InterDigital Communications",
	0x2f7: "Avant Technology",
	0x2f8: "Asrock, Inc.",
	0x2f9: "Availink",
	0x2fa: "Quartics, Inc.",
	0x2fb: "Element CXI",
	0x2fc: "Innovaciones Microelectronicas",
	0x2fd: "VeriSilicon Microelectronics",
	0x2fe: "W5 Networks",
	0x301: "MOVEKING",
	0x302: "Mavrix Technology, Inc.",
	0x303: "CellGuide Ltd.",
	0x304: "Faraday Technology",
	0x305: "Diablo Technologies, Inc.",
	0x306: "Jennic",
	0x307: "Octasic",
	0x308: "Molex Incorporated",
	0x309: "3Leaf Networks",
	0x30a: "Bright Micron Technology",
	0x30b: "Netxen",
	0x30c: "NextWave Broadband Inc.",
	0x30d: "DisplayLink",
	0x30e: "ZMOS Technology",
	0x30f: "Tec-Hill",
	0x310: "Multigig, Inc.",
	0x311: "Amimon",
	0x312: "Euphonic Technologies, Inc.",
	0x313: "BRN Phoenix",
	0x314: "InSilica",
	0x315: "Ember Corporation",
	0x316: "Avexir Technologies Corporation",
	0x317: "Echelon Corporation",
	0x318: "Edgewater Computer Systems",
	0x319: "XMOS Semiconductor Ltd.",
	0x31a: "GENUSION, Inc.",
	0x31b: "Memory Corp NV",
	0x31c: "SiliconBlue Technologies",
	0x31d: "Rambus Inc.",
	0x31e: "Andes Technology Corporation",
	0x31f: "Coronis Systems",
	0x320: "Achronix Semiconductor",
	0x321: "Siano Mobile Silicon Ltd.",
	0x322: "Semtech Corporation",
	0x323: "Pixelworks Inc.",
	0x324: "Gaisler Research AB",
	0x325: "Teranetics",
	0x326: "Toppan Printing Co. Ltd.",
	0x327: "Kingxcon",
	0x328: "Silicon Integrated Systems",
	0x329: "I-O Data Device, Inc.",
	0x32a: "NDS Americas Inc.",
	0x32b: "Solomon Systech Limited",
	0x32c: "On Demand Microelectronics",
	0x32d: "Amicus Wireless Inc.",
	0x32e: "SMARDTV SNC",
	0x32f: "Comsys Communication Ltd.",
	0x330: "Movidia Ltd.",
	0x331: "Javad GNSS, Inc.",
	0x332: "Montage Technology Group",
	0x333: "Trident Microsystems",
	0x334: "Super Talent",
	0x335: "Optichron, Inc.",
	0x336: "Future Waves UK Ltd.",
	0x337: "SiBEAM, Inc.",
	0x338: "Inicore,Inc.",
	0x339: "Virident Systems",
	0x33a: "M2000, Inc.",
	0x33b: "ZeroG Wireless, Inc.",
	0x33c: "Gingle Technology Co. Ltd.",
	0x33d: "Space Micro Inc.",
	0x33e: "Wilocity",
	0x33f: "Novafora, Ic.",
	0x340: "iKoa Corporation",
	0x341: "ASint Technology",
	0x342: "Ramtron",
	0x343: "Plato Networks Inc.",
	0x344: "IPtronics AS",
	0x345: "Infinite-Memories",
	0x346: "Parade Technologies Inc.",
	0x347: "Dune Networks",
	0x348: "GigaDevice Semiconductor",
	0x349: "Modu Ltd.",
	0x34a: "CEITEC",
	0x34b: "Northrop Grumman",
	0x34c: "XRONET Corporation",
	0x34d: "Sicon Semiconductor AB",
	0x34e: "Atla Electronics Co. Ltd.",
	0x34f: "TOPRAM Technology",
	0x350: "Silego Technology Inc.",
	0x351: "Kinglife",
	0x352: "Ability Industries Ltd.",
	0x353: "Silicon Power Computer & Communications",
	0x354: "Augusta Technology, Inc.",
	0x355: "Nantronics Semiconductors",
	0x356: "Hilscher Gesellschaft",
	0x357: "Quixant Ltd.",
	0x358: "Percello Ltd.",
	0x359: "NextIO Inc.",
	0x35a: "Scanimetrics Inc.",
	0x35b: "FS-Semi Company Ltd.",
	0x35c: "Infinera Corporation",
	0x35d: "SandForce Inc.",
	0x35e: "Lexar Media",
	0x35f: "Teradyne Inc.",
	0x360: "Memory Exchange Corp.",
	0x361: "Suzhou Smartek Electronics",
	0x362: "Avantium Corporation",
	0x363: "ATP Electronics Inc.",
	0x364: "Valens Semiconductor Ltd",
	0x365: "Agate Logic, Inc.",
	0x366: "Netronome",
	0x367: "Zenverge, Inc.",
	0x368: "N-trig Ltd",
	0x369: "SanMax Technologies Inc.",
	0x36a: "Contour Semiconductor Inc.",
	0x36b: "TwinMOS",
	0x36c: "Silicon Systems, Inc.",
	0x36d: "V-Color Technology Inc.",
	0x36e: "Certicom Corporation",
	0x36f: "JSC ICC Milandr",
	0x370: "PhotoFast Global Inc.",
	0x371: "InnoDisk Corporation",
	0x372: "Muscle Power",
	0x373: "Energy Micro",
	0x374: "Innofidei",
	0x375: "CopperGate Communications",
	0x376: "Holtek Semiconductor Inc.",
	0x377: "Myson Century, Inc.",
	0x378: "FIDELIX",
	0x379: "Red Digital Cinema",
	0x37a: "Densbits Technology",
	0x37b: "Zempro",
	0x37c: "MoSys",
	0x37d: "Provigent",
	0x37e: "Triad Semiconductor, Inc.",
	0x381: "Siklu Communication Ltd.",
	0x382: "A Force Manufacturing Ltd.",
	0x383: "Strontium",
	0x384: "Abilis Systems",
	0x385: "Siglead, Inc.",
	0x386: "Ubicom, Inc.",
	0x387: "Unifosa Corporation",
	0x388: "Stretch, Inc.",
	0x389: "Lantiq Deutschland GmbH",
	0x38a: "Visipro.",
	0x38b: "EKMemory",
	0x38c: "Microelectronics Institute ZTE",
	0x38d: "Cognovo Ltd.",
	0x38e: "Carry Technology Co. Ltd.",
	0x38f: "Nokia",
	0x390: "King Tiger Technology",
	0x391: "Sierra Wireless",
	0x392: "HT Micron",
	0x393: "Albatron Technology Co. Ltd.",
	0x394: "Leica Geosystems AG",
	0x395: "BroadLight",
	0x396: "AEXEA",
	0x397: "ClariPhy Communications, Inc.",
	0x398: "Green Plug",
	0x399: "Design Art Networks",
	0x39a: "Mach Xtreme Technology Ltd.",
	0x39b: "ATO Solutions Co. Ltd.",
	0x39c: "Ramsta",
	0x39d: "Greenliant Systems, Ltd.",
	0x39e: "Teikon",
	0x39f: "Antec Hadron",
	0x3a0: "NavCom Technology, Inc.",
	0x3a1: "Shanghai Fudan Microelectronics",
	0x3a2: "Calxeda, Inc.",
	0x3a3: "JSC EDC Electronics",
	0x3a4: "Kandit Technology Co. Ltd.",
	0x3a5: "Ramos Technology",
	0x3a6: "Goldenmars Technology",
	0x3a7: "XeL Technology Inc.",
	0x3a8: "Newzone Corporation",
	0x3a9: "ShenZhen MercyPower Tech",
	0x3aa: "Nanjing Yihuo Technology.",
	0x3ab: "Nethra Imaging Inc.",
	0x3ac: "SiTel Semiconductor BV",
	0x3ad: "SolidGear Corporation",
	0x3ae: "Topower Computer Ind Co Ltd.",
	0x3af: "Wilocity",
	0x3b0: "Profichip GmbH",
	0x3b1: "Gerad Technologies",
	0x3b2: "Ritek Corporation",
	0x3b3: "Gomos Technology Limited",
	0x3b4: "Memoright Corporation",
	0x3b5: "D-Broad, Inc.",
	0x3b6: "HiSilicon Technologies",
	0x3b7: "Syndiant Inc..",
	0x3b8: "Enverv Inc.",
	0x3b9: "Cognex",
	0x3ba: "Xinnova Technology Inc.",
	0x3bb: "Ultron AG",
	0x3bc: "Concord Idea Corporation",
	0x3bd: "AIM Corporation",
	0x3be: "Lifetime Memory Products",
	0x3bf: "Ramsway",
	0x3c0: "Recore Systems B.V.",
	0x3c1: "Haotian Jinshibo Science Tech",
	0x3c2: "Being Advanced Memory",
	0x3c3: "Adesto Technologies",
	0x3c4: "Giantec Semiconductor, Inc.",
	0x3c5: "HMD Electronics AG",
	0x3c6: "Gloway International (HK)",
	0x3c7: "Kingcore",
	0x3c8: "Anucell Technology Holding",
	0x3c9: "Accord Software & Systems Pvt. Ltd.",
	0x3ca: "Active-Semi Inc.",
	0x3cb: "Denso Corporation",
}

func mfgNameLookup(mfg uint) string {
	if s, ok := mfgName[mfg]; ok {
		return s
	}
	return "?"
}

// IDCode is a 32-bit JTAG IDCODE.
type IDCode uint32

func (code IDCode) String() string {
	id := uint(code)
	ver := util.GetBits(id, 31, 28)
	part := util.GetBits(id, 27, 12)
	mfg := util.GetBits(id, 11, 1)
	s := []string{}
	s = append(s, fmt.Sprintf("idcode 0x%08x", id))
	s = append(s, fmt.Sprintf("mfg 0x%03x (%s)", mfg, mfgNameLookup(mfg)))
	s = append(s, fmt.Sprintf("part 0x%04x", part))
	s = append(s, fmt.Sprintf("ver 0x%x", ver))
	if id&1 != 1 {
		s = append(s, "leading bit != 1")
	}
	return strings.Join(s, " ")
}

//-----------------------------------------------------------------------------
