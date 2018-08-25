#version 330

uniform sampler2D tex;
uniform vec4 col_tint;
uniform float points;

in vec2 tex_coord;
in float index;

out vec4 color;

void main() {
    vec4 in_color = texture(tex, tex_coord);
	color = in_color * col_tint * vec4(1, 1, 1, 1-smoothstep(points / 3, points, index));
}