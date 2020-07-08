#version 330

uniform sampler2DArray tex;
uniform vec4 col_tint;
uniform float points;
//uniform float scale;

in vec2 tex_coord;
in float index;

out vec4 color;

void main() {
    vec4 in_color = texture(tex, vec3(tex_coord, 0));
    //float scaling = 0.01 + 0.99 * 2 * length(tex_coord - vec2(0.5, 0.5));
	color = in_color * col_tint * vec4(1.0, 1.0, 1.0, 1-smoothstep(points / 3.0, points, index));
    //color.rgb *= mix(1 - 0.99 * scale / 11, 1, index / points);
}