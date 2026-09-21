package officials

// currentSenators was transcribed from the U.S. Senate's current-members
// roster on 2026-09-21. The data remains dated and USCIS verification remains
// required in the interface.
var currentSenators = map[string][]string{
	"AL": {"Katie Boyd Britt", "Tommy Tuberville"}, "AK": {"Lisa Murkowski", "Dan Sullivan"}, "AZ": {"Ruben Gallego", "Mark Kelly"}, "AR": {"John Boozman", "Tom Cotton"},
	"CA": {"Alex Padilla", "Adam B. Schiff"}, "CO": {"Michael F. Bennet", "John W. Hickenlooper"}, "CT": {"Richard Blumenthal", "Christopher Murphy"}, "DE": {"Lisa Blunt Rochester", "Christopher A. Coons"},
	"FL": {"Ashley Moody", "Rick Scott"}, "GA": {"Jon Ossoff", "Raphael G. Warnock"}, "HI": {"Mazie K. Hirono", "Brian Schatz"}, "ID": {"Mike Crapo", "James E. Risch"},
	"IL": {"Tammy Duckworth", "Richard J. Durbin"}, "IN": {"Jim Banks", "Todd Young"}, "IA": {"Joni Ernst", "Chuck Grassley"}, "KS": {"Roger Marshall", "Jerry Moran"},
	"KY": {"Mitch McConnell", "Rand Paul"}, "LA": {"Bill Cassidy", "John Kennedy"}, "ME": {"Susan M. Collins", "Angus S. King, Jr."}, "MD": {"Angela D. Alsobrooks", "Chris Van Hollen"},
	"MA": {"Edward J. Markey", "Elizabeth Warren"}, "MI": {"Gary C. Peters", "Elissa Slotkin"}, "MN": {"Amy Klobuchar", "Tina Smith"}, "MS": {"Cindy Hyde-Smith", "Roger F. Wicker"},
	"MO": {"Josh Hawley", "Eric Schmitt"}, "MT": {"Steve Daines", "Tim Sheehy"}, "NE": {"Deb Fischer", "Pete Ricketts"}, "NV": {"Catherine Cortez Masto", "Jacky Rosen"},
	"NH": {"Margaret Wood Hassan", "Jeanne Shaheen"}, "NJ": {"Cory A. Booker", "Andy Kim"}, "NM": {"Martin Heinrich", "Ben Ray Luján"}, "NY": {"Kirsten E. Gillibrand", "Charles E. Schumer"},
	"NC": {"Ted Budd", "Thom Tillis"}, "ND": {"Kevin Cramer", "John Hoeven"}, "OH": {"Jon Husted", "Bernie Moreno"}, "OK": {"Alan Armstrong", "James Lankford"},
	"OR": {"Jeff Merkley", "Ron Wyden"}, "PA": {"John Fetterman", "David McCormick"}, "RI": {"Jack Reed", "Sheldon Whitehouse"}, "SC": {"Darline Graham", "Tim Scott"},
	"SD": {"Mike Rounds", "John Thune"}, "TN": {"Marsha Blackburn", "Bill Hagerty"}, "TX": {"John Cornyn", "Ted Cruz"}, "UT": {"John R. Curtis", "Mike Lee"},
	"VT": {"Bernard Sanders", "Peter Welch"}, "VA": {"Tim Kaine", "Mark R. Warner"}, "WA": {"Maria Cantwell", "Patty Murray"}, "WV": {"Shelley Moore Capito", "James C. Justice"},
	"WI": {"Tammy Baldwin", "Ron Johnson"}, "WY": {"John Barrasso", "Cynthia M. Lummis"}, "DC": {},
}
