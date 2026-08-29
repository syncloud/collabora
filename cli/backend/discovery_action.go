package backend

type discoveryAction struct {
	Name   string `xml:"name,attr"`
	Ext    string `xml:"ext,attr"`
	URLSrc string `xml:"urlsrc,attr"`
}
