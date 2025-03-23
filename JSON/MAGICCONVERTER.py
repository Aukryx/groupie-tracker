import json
# Lire les coordonnées depuis le fichier JSON
with open('coordinates.json', 'r') as f:
    coordinates = json.load(f)
# Créer une liste de fonctionnalités GeoJSON
features = []
for location, coord in coordinates.items():
    feature = {
        "type": "Feature",
        "geometry": {
            "type": "Point",
            "coordinates": coord
        },
        "properties": {
            "name": location
        }
    }
    features.append(feature)
# Créer l'objet GeoJSON final
geojson = {
    "type": "FeatureCollection",
    "features": features
}
# Écrire l'objet GeoJSON dans un fichier
with open('data.geojson', 'w') as f:
    json.dump(geojson, f, indent=2)
print("Le fichier locations.geojson a été généré avec succès.")