// track/track.go
package track

var names = map[int8]string{
    0:  "Melbourne",
    2:  "Shanghai",
    3:  "Sakhir",
    4:  "Catalunya",
    5:  "Monaco",
    6:  "Montreal",
    7:  "Silverstone",
    9:  "Hungaroring",
    10: "Spa",
    11: "Monza",
    12: "Singapore",
    13: "Suzuka",
    14: "Abu Dhabi",
    15: "Texas",
    16: "Brazil",
    17: "Austria",
    19: "Mexico",
    20: "Baku",
    26: "Zandvoort",
    27: "Imola",
    29: "Jeddah",
    30: "Miami",
    31: "Las Vegas",
    32: "Losail",
    39: "Silverstone (Reverse)",
    40: "Austria (Reverse)",
    41: "Zandvoort (Reverse)",
}

func Name(id int8) string {
    if name, ok := names[id]; ok {
        return name
    }
    return "Unknown"
}