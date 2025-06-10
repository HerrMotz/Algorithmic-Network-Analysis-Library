input_file = "fb-pages-public-figure.edges"
output_file = "fb-pages-public-figure2.edges"

with open(input_file, "r") as infile, open(output_file, "w") as outfile:
    for line in infile:
        # Entferne das Komma und schreibe die Zeile in die Ausgabedatei
        outfile.write(line.replace(",", " ").strip() + "\n")