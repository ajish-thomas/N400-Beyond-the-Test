package quiz

// semanticRule records a spoken-style phrase whose meaning is stated in an
// official USCIS source but whose wording differs from the 128-question
// answer bullet. Source identifies the checked-in text edition used in review.
type semanticRule struct {
	QuestionID  int
	Terms       []string
	AnswerIndex int
	Source      string
}

var semanticRules = []semanticRule{
	{3, []string{"set", "up", "government"}, 0, "Study Guide, Chapter 1: “The first three sections of the Constitution set up the U.S. government.”"},
	{8, []string{"announced", "america", "independence", "britain"}, 0, "Study Guide, Chapter 8: “The Declaration of Independence announced our independence from Great Britain.”"},
	{18, []string{"congress", "make", "federal", "law"}, 0, "Study Guide, Chapter 2: “Congress makes federal laws.”"},
	{20, []string{"congress", "make", "federal", "law"}, 0, "Study Guide, Chapter 2: “Congress makes federal laws.”"},
	{47, []string{"cabinet", "group", "advisor"}, 0, "Study Guide, Chapter 3: “The President’s Cabinet advises the President.”"},
	{84, []string{"federalist", "paper", "supported", "passing", "constitution"}, 1, "Study Guide, Chapter 5: “The Federalist Papers supported passing the U.S. Constitution.”"},
	{95, []string{"enslaved", "southern", "state", "free"}, 3, "Study Guide, Chapter 10: “The Emancipation Proclamation said the enslaved people in the Southern states were free.”"},
	{106, []string{"japan", "attack", "pearl", "harbor"}, 1, "Study Guide, Chapter 11: “Japan attacked the United States at a U.S. naval base in Hawaii called Pearl Harbor.”"},
	{112, []string{"end", "racial", "discrimination"}, 0, "Study Guide, Chapter 11: “The movement to end racial discrimination in the United States is called the Civil Rights Movement.”"},
}
