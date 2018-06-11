#version 330

uniform sampler2D tex;
uniform vec4 col_border;

in vec2 tex_coord;
out vec4 color;
void main()
{
    vec4 in_color = texture2D(tex, tex_coord);

	//vec4 col_tint = vec4(0, 0, 1, 0.5);

	color = in_color*col_border;//vec4(mix(in_color.xyz*col_border.xyz, col_tint.xyz, 1.0-in_color.a), in_color.a*col_border.a);
}