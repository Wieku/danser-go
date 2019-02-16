#version 330

in vec4 col_tint;
in vec3 tex_coord;
in float additive;

uniform sampler2DArray tex;

out vec4 color;

void main()
{
    vec4 in_color = texture(tex, tex_coord);
	color = in_color*col_tint;
	color.rgb *= color.a;
	color.a *= additive;
}