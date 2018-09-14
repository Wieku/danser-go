#version 330

in vec4 col_tint;
in vec2 tex_coord;
in float additive;

uniform sampler2D tex;

out vec4 color;

void main()
{
    vec4 in_color = texture(tex, tex_coord);
	color = in_color*col_tint;
	color.rgb *= color.a;
	if (additive == 1) {
	    color.a = 0;
	}
}